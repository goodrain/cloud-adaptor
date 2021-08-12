// RAINBOND, Application Management Platform
// Copyright (C) 2020-2020 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package repo

import (
	"github.com/pkg/errors"
	"goodrain.com/cloud-adaptor/internal/model"
	"goodrain.com/cloud-adaptor/pkg/util/uuidutil"
	"gorm.io/gorm"
)

// TaskEventRepo enterprise task event
type TaskEventRepo struct {
	DB *gorm.DB `inject:""`
}

// NewTaskEventRepo new Enterprise repoo
func NewTaskEventRepo(db *gorm.DB) TaskEventRepository {
	return &TaskEventRepo{DB: db}
}

// Transaction -
func (t *TaskEventRepo) Transaction(tx *gorm.DB) TaskEventRepository {
	return &TaskEventRepo{DB: tx}
}

//Create create an event
func (t *TaskEventRepo) Create(te *model.TaskEvent) error {
	if len(te.Message) > 512 {
		te.Message = te.Message[:512]
	}
	var old model.TaskEvent
	if err := t.DB.Where("eid = ? and task_id=? and step_type=?", te.EnterpriseID, te.TaskID, te.StepType).Take(&old).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// not found error, create new
			if te.EventID == "" {
				te.EventID = uuidutil.NewUUID()
			}
			if err := t.DB.Save(te).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}
	//Prevent successful events from being overwritten
	if old.Status != "success" {
		old.Message = te.Message
		old.Status = te.Status
		old.Reason = te.Reason
		return t.DB.Save(&old).Error
	}
	return nil
}

//ListEvent list task events
func (t *TaskEventRepo) ListEvent(eid, taskID string) ([]*model.TaskEvent, error) {
	var list []*model.TaskEvent
	if err := t.DB.Where("eid = ? and task_id=?", eid, taskID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (t *TaskEventRepo) UpdateStatusInBatch(eventIDs []string, status string) error {
	if err := t.DB.Where("event_id in (?)", eventIDs).Updates(&model.TaskEvent{Status: status}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
