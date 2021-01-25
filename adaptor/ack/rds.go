package ack

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/adaptor/v1alpha1"
)

func (a *ackAdaptor) SupportDB() bool {
	return true
}
func (a *ackAdaptor) DescribeDBInstance(clusterID, regionID, ZoneID string) (*rds.DBInstanceInDescribeDBInstances, error) {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := rds.CreateDescribeDBInstancesRequest()
	request.RegionId = regionID
	request.ZoneId = ZoneID
	res, _ := ecsclient.DescribeDBInstances(request)
	if res != nil && len(res.Items.DBInstance) > 0 {
		for _, instance := range res.Items.DBInstance {
			logrus.Infof("db instance %s", instance.DBInstanceDescription)
			if instance.DBInstanceDescription == "rainbond-region-db_"+clusterID {
				return &instance, nil
			}
		}
	}
	return nil, nil
}

func (a *ackAdaptor) DescribeDBInstanceNetInfo(regionID, instanceID string) (response *rds.DescribeDBInstanceNetInfoResponse, err error) {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := rds.CreateDescribeDBInstanceNetInfoRequest()
	request.RegionId = regionID
	request.DBInstanceId = instanceID
	return ecsclient.DescribeDBInstanceNetInfo(request)
}

func (a *ackAdaptor) CreateDBInstance(clusterID, regionID, ZoneID, VPCId, VSwitchID, podCIDR string) (*rds.CreateDBInstanceResponse, error) {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return nil, err
	}
	request := rds.CreateCreateDBInstanceRequest()
	request.RegionId = regionID
	request.Engine = "MySQL"
	request.EngineVersion = "8.0"
	request.DBInstanceClass = "rds.mysql.s1.small" //高可用版 入门级 通用型 1core 2GB（单机基础版）
	request.DBInstanceStorage = requests.NewInteger(5)
	request.InstanceNetworkType = "VPC"
	request.DBInstanceNetType = "Intranet" //内网连接
	request.VPCId = VPCId
	request.ZoneId = ZoneID
	request.VSwitchId = VSwitchID
	securitys := []string{podCIDR}
	vpc, _ := a.DescribeVPC(regionID, VPCId)
	if vpc != nil {
		securitys = append(securitys, vpc.CidrBlock)
	} else {
		securitys = append(securitys, "10.22.0.0/16")
	}
	request.SecurityIPList = strings.Join(securitys, ",")
	logrus.Infof("Security ip list: %s", request.SecurityIPList)
	request.PayType = "Postpaid" //后付费
	request.DBInstanceDescription = "rainbond-region-db_" + clusterID
	res, err := ecsclient.CreateDBInstance(request)
	if err != nil {
		return nil, err
	}
	if err := a.WaitingDBInstanceReady(context.Background(), regionID, res.DBInstanceId); err != nil {
		return nil, err
	}
	return res, nil
}
func (a *ackAdaptor) WaitingDBInstanceReady(ctx context.Context, regionID, dbInstanceID string) error {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	req := rds.CreateDescribeDBInstancesRequest()
	req.Scheme = "https"
	req.DBInstanceId = dbInstanceID
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	timer := time.NewTimer(time.Minute * 10)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			return fmt.Errorf("create acount timeout")
		case <-ticker.C:
		}
		res, err := ecsclient.DescribeDBInstances(req)
		if err != nil {
			return err
		}
		for _, in := range res.Items.DBInstance {
			if in.DBInstanceStatus == "Running" {
				return nil
			}
			logrus.Infof("db %s status is %s", in.DBInstanceDescription, in.DBInstanceStatus)
		}
	}
}
func (a *ackAdaptor) createDBAcount(regionID, instanceID, name, password string) error {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	req := rds.CreateDescribeAccountsRequest()
	req.Scheme = "https"
	req.AccountName = name
	req.DBInstanceId = instanceID
	res, err := ecsclient.DescribeAccounts(req)
	if err != nil && !strings.Contains(err.Error(), "The specified resource does not exist") {
		return err
	}
	for _, account := range res.Accounts.DBInstanceAccount {
		if account.AccountStatus == "Available" && account.AccountName == name {
			logrus.Infof("db account %s status is %s", name, account.AccountStatus)
			return nil
		}
	}
	acount := rds.CreateCreateAccountRequest()
	acount.DBInstanceId = instanceID
	acount.AccountName = name
	acount.AccountPassword = password
	acount.AccountDescription = "create by rainbond cloud"
	response, err := ecsclient.CreateAccount(acount)
	if err != nil {
		return fmt.Errorf("create rds(mysql) user from alibaba api failure:%s", response.String())
	}
	if !response.IsSuccess() {
		return fmt.Errorf("create rds(mysql) user from alibaba api failure:%s", response.String())
	}
	ticker := time.NewTicker(time.Second * 1)
	timer := time.NewTimer(time.Minute * 3)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("create acount timeout")
		}
		req := rds.CreateDescribeAccountsRequest()
		req.Scheme = "https"
		req.AccountName = name
		req.DBInstanceId = instanceID
		res, err := ecsclient.DescribeAccounts(req)
		if err != nil {
			return err
		}
		for _, account := range res.Accounts.DBInstanceAccount {
			logrus.Infof("db account %s status is %s", name, account.AccountStatus)
			if account.AccountStatus == "Available" && account.AccountName == name {
				return nil
			}
		}
	}
}

func (a *ackAdaptor) createDatabase(regionID, instanceID, dbName string) error {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	req := rds.CreateDescribeDatabasesRequest()
	req.Scheme = "https"
	req.DBName = dbName
	req.DBInstanceId = instanceID
	dres, err := ecsclient.DescribeDatabases(req)
	if err != nil && !strings.Contains(err.Error(), "The specified resource does not exist") {
		return err
	}
	for _, db := range dres.Databases.Database {
		logrus.Infof("db database %s status is %s", dbName, db.DBStatus)
		if db.DBStatus == "Running" && db.DBName == dbName {
			return nil
		}
	}
	database := rds.CreateCreateDatabaseRequest()
	database.DBInstanceId = instanceID
	database.DBName = dbName
	database.CharacterSetName = "utf8"
	database.DBDescription = "create by rainbond cloud"
	res, err := ecsclient.CreateDatabase(database)
	if err != nil {
		return fmt.Errorf("create rds(mysql) database from alibaba api failure:%s", res.String())
	}
	if !res.IsSuccess() {
		return fmt.Errorf("create rds(mysql) database from alibaba api failure:%s", res.String())
	}
	ticker := time.NewTicker(time.Second * 1)
	timer := time.NewTimer(time.Minute * 3)
	defer timer.Stop()
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-timer.C:
			return fmt.Errorf("create database timeout")
		}
		req := rds.CreateDescribeDatabasesRequest()
		req.Scheme = "https"
		req.DBName = dbName
		req.DBInstanceId = instanceID
		res, err := ecsclient.DescribeDatabases(req)
		if err != nil {
			return err
		}
		for _, db := range res.Databases.Database {
			logrus.Infof("db database %s status is %s", dbName, db.DBStatus)
			if db.DBStatus == "Running" && db.DBName == dbName {
				return nil
			}
		}
	}
}

func (a *ackAdaptor) createGrantAccountPrivilege(regionID, instanceID, dbName, userName string) error {
	ecsclient, err := rds.NewClientWithAccessKey(regionID, a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return err
	}
	privilege := rds.CreateGrantAccountPrivilegeRequest()
	privilege.DBInstanceId = instanceID
	privilege.AccountName = userName
	privilege.AccountPrivilege = "ReadWrite"
	privilege.DBName = dbName
	resp, err := ecsclient.GrantAccountPrivilege(privilege)
	if err != nil {
		return fmt.Errorf("create rds(mysql) user privilege from alibaba api failure:%s", resp.String())
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("create rds(mysql) user privilege from alibaba api failure:%s", resp.String())
	}
	return nil
}
func (a *ackAdaptor) CreateDB(db *v1alpha1.Database) error {
	if db.RegionID == "" {
		return fmt.Errorf("not privide region id")
	}
	//create instance
	if db.InstanceID == "" {
		instance, err := a.DescribeDBInstance(db.ClusterID, db.RegionID, db.ZoneID)
		if err != nil {
			return fmt.Errorf("describe rds(mysql) from alibaba api failure:%s", err.Error())
		}
		if instance != nil {
			db.InstanceID = instance.DBInstanceId
			res, err := a.DescribeDBInstanceNetInfo(db.RegionID, instance.DBInstanceId)
			if err != nil {
				return fmt.Errorf("describe rds(mysql) net info from alibaba api failure:%s", err.Error())
			}
			for _, addr := range res.DBInstanceNetInfos.DBInstanceNetInfo {
				if addr.IPType == "Private" {
					db.Host = addr.ConnectionString
					db.Port, _ = strconv.Atoi(addr.Port)
				}
			}
			// if there is no private connection, use Public
			if db.Host == "" {
				for _, addr := range res.DBInstanceNetInfos.DBInstanceNetInfo {
					if addr.IPType == "Public" {
						db.Host = addr.ConnectionString
						db.Port, _ = strconv.Atoi(addr.Port)
					}
				}
			}
		} else {
			response, err := a.CreateDBInstance(db.ClusterID, db.RegionID, db.ZoneID, db.VPCID, db.VSwitchID, db.PodCIDR)
			if err != nil {
				return fmt.Errorf("create rds(mysql) from alibaba api failure:%s", err.Error())
			}
			db.InstanceID = response.DBInstanceId
			db.Host = response.ConnectionString
			db.Port, _ = strconv.Atoi(response.Port)
		}
	}

	//create acount
	if err := a.createDBAcount(db.RegionID, db.InstanceID, db.UserName, db.Password); err != nil {
		return fmt.Errorf("create rds(mysql) account failure:%s", err.Error())
	}
	//create database
	if err := a.createDatabase(db.RegionID, db.InstanceID, db.Name); err != nil {
		return fmt.Errorf("create rds(mysql) database from alibaba api failure:%s", err.Error())
	}
	//grant account privilege
	if err := a.createGrantAccountPrivilege(db.RegionID, db.InstanceID, db.Name, db.UserName); err != nil {
		return fmt.Errorf("create rds(mysql) user privilege from alibaba api failure:%s", err.Error())
	}
	return nil
}
