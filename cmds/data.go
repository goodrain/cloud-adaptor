package cmds

import (
	"encoding/json"
	"io/ioutil"

	cli "github.com/urfave/cli/v2"
	"goodrain.com/cloud-adaptor/cmd/cloud-adaptor/config"
	"goodrain.com/cloud-adaptor/internal/data/model"
	"goodrain.com/cloud-adaptor/pkg/infrastructure/datastore"
)

var dataCommand = &cli.Command{
	Name: "data",
	Subcommands: []*cli.Command{
		{
			Name: "export",
			Flags: append(dbInfoFlag,
				&cli.StringFlag{
					Name:     "fileName",
					Required: true,
					Usage:    "export file name",
				},
			),
			Action: exportData,
			Usage:  "export adaptor all data",
		},
		{
			Name: "import",
			Flags: append(dbInfoFlag,
				&cli.StringFlag{
					Name:     "fileName",
					Required: true,
					Usage:    "import file name",
				},
			),
			Action: importData,
			Usage:  "import adaptor all data",
		},
	},
}

func exportData(ctx *cli.Context) error {
	config.Parse(ctx)
	db := datastore.NewDB()
	defer func() { _ = db.Close() }()

	var result model.BackupListModelData
	db.Table(db.NewScope(&model.CloudAccessKey{}).TableName()).Scan(&result.CloudAccessKeys)
	db.Table(db.NewScope(&model.CreateKubernetesTask{}).TableName()).Scan(&result.CreateKubernetesTasks)
	db.Table(db.NewScope(&model.InitRainbondTask{}).TableName()).Scan(&result.InitRainbondTasks)
	db.Table(db.NewScope(&model.TaskEvent{}).TableName()).Scan(&result.TaskEvents)

	data, err := json.Marshal(result)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := writeFile(ctx.String("fileName"), data); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func importData(ctx *cli.Context) error {
	config.Parse(ctx)
	db := datastore.NewDB()
	defer func() { _ = db.Close() }()

	bytes, err := readFile(ctx.String("fileName"))
	if err != nil {
		return cli.Exit(err, 1)
	}

	var data model.BackupListModelData
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return cli.Exit(err, 1)
	}

	for _, accessKey := range data.CloudAccessKeys {
		if err := db.Create(&accessKey).Error; err != nil {
			return cli.Exit(err, 1)
		}
	}
	for _, createTask := range data.CreateKubernetesTasks {
		if err := db.Create(&createTask).Error; err != nil {
			return cli.Exit(err, 1)
		}
	}
	for _, initTask := range data.InitRainbondTasks {
		if err := db.Create(&initTask).Error; err != nil {
			return cli.Exit(err, 1)
		}
	}
	for _, taskEvent := range data.TaskEvents {
		if err := db.Create(&taskEvent).Error; err != nil {
			return cli.Exit(err, 1)
		}
	}
	return nil
}

func writeFile(fileName string, data []byte) error {
	err := ioutil.WriteFile("./data/"+fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readFile(fileName string) ([]byte, error) {
	bytes, err := ioutil.ReadFile("./data/" + fileName)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
