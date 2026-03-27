package service

import (
	"errors"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth/user"
	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	adminModels "go-admin/app/admin/models"
	platformModels "go-admin/app/platform/models"
	"go-admin/app/platform/service/dto"
	"go-admin/common/authctx"
)

type Workflow struct {
	service.Service
}

type WorkflowDefinitionDetail struct {
	Definition platformModels.WorkflowDefinition       `json:"definition"`
	Nodes      []platformModels.WorkflowDefinitionNode `json:"nodes"`
}

type WorkflowInstanceDetail struct {
	Instance platformModels.WorkflowInstance         `json:"instance"`
	Binding  platformModels.WorkflowBusinessBinding  `json:"binding"`
	Tasks    []platformModels.WorkflowTask           `json:"tasks"`
	Actions  []platformModels.WorkflowActionLog      `json:"actions"`
	Nodes    []platformModels.WorkflowDefinitionNode `json:"nodes"`
}

type WorkflowTodoTaskItem struct {
	TaskId          int        `json:"taskId"`
	InstanceId      int        `json:"instanceId"`
	DefinitionId    int        `json:"definitionId"`
	DefinitionName  string     `json:"definitionName"`
	ModuleKey       string     `json:"moduleKey"`
	BusinessType    string     `json:"businessType"`
	BusinessId      string     `json:"businessId"`
	BusinessNo      string     `json:"businessNo"`
	Title           string     `json:"title"`
	StarterId       int        `json:"starterId"`
	StarterName     string     `json:"starterName"`
	Status          string     `json:"status"`
	CurrentNodeId   int        `json:"currentNodeId"`
	CurrentNodeKey  string     `json:"currentNodeKey"`
	CurrentNodeName string     `json:"currentNodeName"`
	TaskStatus      string     `json:"taskStatus"`
	AssigneeType    string     `json:"assigneeType"`
	AssigneeId      int        `json:"assigneeId"`
	AssigneeName    string     `json:"assigneeName"`
	TaskCreatedAt   time.Time  `json:"taskCreatedAt"`
	StartedAt       time.Time  `json:"startedAt"`
	FinishedAt      *time.Time `json:"finishedAt"`
}

type WorkflowStartedInstanceItem struct {
	InstanceId       int        `json:"instanceId"`
	DefinitionId     int        `json:"definitionId"`
	DefinitionName   string     `json:"definitionName"`
	ModuleKey        string     `json:"moduleKey"`
	BusinessType     string     `json:"businessType"`
	BusinessId       string     `json:"businessId"`
	BusinessNo       string     `json:"businessNo"`
	Title            string     `json:"title"`
	StarterId        int        `json:"starterId"`
	StarterName      string     `json:"starterName"`
	Status           string     `json:"status"`
	CurrentNodeId    int        `json:"currentNodeId"`
	CurrentNodeKey   string     `json:"currentNodeKey"`
	CurrentNodeName  string     `json:"currentNodeName"`
	LastAction       string     `json:"lastAction"`
	LastActionRemark string     `json:"lastActionRemark"`
	StartedAt        time.Time  `json:"startedAt"`
	FinishedAt       *time.Time `json:"finishedAt"`
}

func (e *Workflow) GetDefinitionPage(c *dto.WorkflowDefinitionGetPageReq, list *[]platformModels.WorkflowDefinition, count *int64) error {
	db := e.Orm.Model(&platformModels.WorkflowDefinition{})
	if c.DefinitionName != "" {
		db = db.Where("definition_name LIKE ?", "%"+strings.TrimSpace(c.DefinitionName)+"%")
	}
	if c.ModuleKey != "" {
		db = db.Where("module_key = ?", strings.TrimSpace(c.ModuleKey))
	}
	if c.BusinessType != "" {
		db = db.Where("business_type = ?", strings.TrimSpace(c.BusinessType))
	}
	if c.Status != "" {
		db = db.Where("status = ?", c.Status)
	}
	return db.Order("updated_at DESC, definition_id DESC").
		Offset((c.GetPageIndex() - 1) * c.GetPageSize()).
		Limit(c.GetPageSize()).
		Find(list).Limit(-1).Offset(-1).Count(count).Error
}

func (e *Workflow) GetTodoTaskPage(c *dto.WorkflowTodoTaskGetPageReq, list *[]WorkflowTodoTaskItem, count *int64, currentUserID int, currentRoleIDs []int) error {
	db := e.Orm.Table("wf_task AS t").
		Select([]string{
			"t.task_id",
			"t.instance_id",
			"t.definition_id",
			"i.definition_name",
			"i.module_key",
			"i.business_type",
			"i.business_id",
			"i.business_no",
			"i.title",
			"i.starter_id",
			"i.starter_name",
			"i.status",
			"i.current_node_id",
			"i.current_node_key",
			"i.current_node_name",
			"t.status AS task_status",
			"t.assignee_type",
			"t.assignee_id",
			"t.assignee_name",
			"t.created_at AS task_created_at",
			"i.started_at",
			"i.finished_at",
		}).
		Joins("LEFT JOIN wf_instance AS i ON i.instance_id = t.instance_id").
		Where("t.status = ?", dto.WorkflowTaskPending)

	if len(currentRoleIDs) > 0 {
		db = db.Where("(t.assignee_type = ? AND t.assignee_id = ?) OR (t.assignee_type = ? AND t.assignee_id IN ?)",
			dto.WorkflowApproverUser, currentUserID, dto.WorkflowApproverRole, currentRoleIDs)
	} else {
		db = db.Where("t.assignee_type = ? AND t.assignee_id = ?", dto.WorkflowApproverUser, currentUserID)
	}

	if c.Title != "" {
		db = db.Where("i.title LIKE ?", "%"+strings.TrimSpace(c.Title)+"%")
	}
	if c.BusinessType != "" {
		db = db.Where("i.business_type = ?", strings.TrimSpace(c.BusinessType))
	}
	if c.BusinessNo != "" {
		db = db.Where("i.business_no LIKE ?", "%"+strings.TrimSpace(c.BusinessNo)+"%")
	}
	if c.Status != "" {
		db = db.Where("i.status = ?", strings.TrimSpace(c.Status))
	}

	if err := db.Session(&gorm.Session{}).Count(count).Error; err != nil {
		return err
	}

	return db.Order("t.created_at DESC, t.task_id DESC").
		Offset((c.GetPageIndex() - 1) * c.GetPageSize()).
		Limit(c.GetPageSize()).
		Scan(list).Error
}

func (e *Workflow) GetStartedInstancePage(c *dto.WorkflowStartedInstanceGetPageReq, list *[]WorkflowStartedInstanceItem, count *int64, currentUserID int) error {
	db := e.Orm.Table("wf_instance AS i").
		Select([]string{
			"i.instance_id",
			"i.definition_id",
			"i.definition_name",
			"i.module_key",
			"i.business_type",
			"i.business_id",
			"i.business_no",
			"i.title",
			"i.starter_id",
			"i.starter_name",
			"i.status",
			"i.current_node_id",
			"i.current_node_key",
			"i.current_node_name",
			"i.last_action",
			"i.last_action_remark",
			"i.started_at",
			"i.finished_at",
		}).
		Where("i.starter_id = ?", currentUserID)

	if c.Title != "" {
		db = db.Where("i.title LIKE ?", "%"+strings.TrimSpace(c.Title)+"%")
	}
	if c.BusinessType != "" {
		db = db.Where("i.business_type = ?", strings.TrimSpace(c.BusinessType))
	}
	if c.BusinessNo != "" {
		db = db.Where("i.business_no LIKE ?", "%"+strings.TrimSpace(c.BusinessNo)+"%")
	}
	if c.Status != "" {
		db = db.Where("i.status = ?", strings.TrimSpace(c.Status))
	}

	if err := db.Session(&gorm.Session{}).Count(count).Error; err != nil {
		return err
	}

	return db.Order("i.started_at DESC, i.instance_id DESC").
		Offset((c.GetPageIndex() - 1) * c.GetPageSize()).
		Limit(c.GetPageSize()).
		Scan(list).Error
}

func (e *Workflow) GetDefinition(id int) (*WorkflowDefinitionDetail, error) {
	var definition platformModels.WorkflowDefinition
	if err := e.Orm.First(&definition, id).Error; err != nil {
		return nil, err
	}
	nodes := make([]platformModels.WorkflowDefinitionNode, 0)
	if err := e.Orm.Where("definition_id = ?", id).Order("sort ASC, node_id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return &WorkflowDefinitionDetail{
		Definition: definition,
		Nodes:      nodes,
	}, nil
}

func (e *Workflow) InsertDefinition(c *dto.WorkflowDefinitionInsertReq) (platformModels.WorkflowDefinition, error) {
	c.Normalize()
	var module platformModels.ModuleRegistry
	if err := e.Orm.Where("module_key = ? AND status = ?", c.ModuleKey, "2").First(&module).Error; err != nil {
		return platformModels.WorkflowDefinition{}, errors.New("模块未注册或未启用")
	}
	var count int64
	if err := e.Orm.Model(&platformModels.WorkflowDefinition{}).Where("definition_key = ?", c.DefinitionKey).Count(&count).Error; err != nil {
		return platformModels.WorkflowDefinition{}, err
	}
	if count > 0 {
		return platformModels.WorkflowDefinition{}, errors.New("流程定义编码已存在")
	}
	var model platformModels.WorkflowDefinition
	c.Generate(&model)
	model.CreateBy = c.CreateBy
	model.UpdateBy = c.CreateBy
	return model, e.Orm.Create(&model).Error
}

func (e *Workflow) UpdateDefinition(c *dto.WorkflowDefinitionUpdateReq) (platformModels.WorkflowDefinition, error) {
	c.Normalize()
	var module platformModels.ModuleRegistry
	if err := e.Orm.Where("module_key = ? AND status = ?", c.ModuleKey, "2").First(&module).Error; err != nil {
		return platformModels.WorkflowDefinition{}, errors.New("模块未注册或未启用")
	}
	var count int64
	if err := e.Orm.Model(&platformModels.WorkflowDefinition{}).
		Where("definition_key = ? AND definition_id <> ?", c.DefinitionKey, c.DefinitionId).
		Count(&count).Error; err != nil {
		return platformModels.WorkflowDefinition{}, err
	}
	if count > 0 {
		return platformModels.WorkflowDefinition{}, errors.New("流程定义编码已存在")
	}
	var model platformModels.WorkflowDefinition
	if err := e.Orm.First(&model, c.DefinitionId).Error; err != nil {
		return platformModels.WorkflowDefinition{}, err
	}
	c.Generate(&model)
	model.UpdateBy = c.UpdateBy
	return model, e.Orm.Save(&model).Error
}

func (e *Workflow) DeleteDefinition(id int) error {
	var instanceCount int64
	if err := e.Orm.Model(&platformModels.WorkflowInstance{}).Where("definition_id = ?", id).Count(&instanceCount).Error; err != nil {
		return err
	}
	if instanceCount > 0 {
		return errors.New("已有流程实例，不能删除流程定义")
	}
	return e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("definition_id = ?", id).Delete(&platformModels.WorkflowDefinitionNode{}).Error; err != nil {
			return err
		}
		return tx.Delete(&platformModels.WorkflowDefinition{}, id).Error
	})
}

func (e *Workflow) GetDefinitionNodes(definitionID int, list *[]platformModels.WorkflowDefinitionNode) error {
	return e.Orm.Where("definition_id = ?", definitionID).Order("sort ASC, node_id ASC").Find(list).Error
}

func (e *Workflow) SaveDefinitionNodes(definitionID int, req *dto.WorkflowDefinitionNodeSaveReq) error {
	var definition platformModels.WorkflowDefinition
	if err := e.Orm.First(&definition, definitionID).Error; err != nil {
		return err
	}
	if len(req.Nodes) == 0 {
		return errors.New("节点列表不能为空")
	}

	startCount := 0
	endCount := 0
	approveCount := 0
	sortSet := make(map[int]struct{}, len(req.Nodes))
	keySet := make(map[string]struct{}, len(req.Nodes))
	nodeModels := make([]platformModels.WorkflowDefinitionNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		node.Normalize()
		if err := node.Validate(); err != nil {
			return err
		}
		if _, ok := sortSet[node.Sort]; ok {
			return errors.New("节点排序不能重复")
		}
		sortSet[node.Sort] = struct{}{}
		if _, ok := keySet[node.NodeKey]; ok {
			return errors.New("节点编码不能重复")
		}
		keySet[node.NodeKey] = struct{}{}

		switch node.NodeType {
		case dto.WorkflowNodeTypeStart:
			startCount++
		case dto.WorkflowNodeTypeEnd:
			endCount++
		case dto.WorkflowNodeTypeApprove:
			approveCount++
		}

		nodeModel := platformModels.WorkflowDefinitionNode{
			NodeId:        node.NodeId,
			DefinitionId:  definitionID,
			NodeKey:       node.NodeKey,
			NodeName:      node.NodeName,
			NodeType:      node.NodeType,
			Sort:          node.Sort,
			ApproverType:  node.ApproverType,
			ApproverValue: node.ApproverValue,
			ApproverName:  node.ApproverName,
			Remark:        node.Remark,
		}
		nodeModel.CreateBy = req.CreateBy
		nodeModel.UpdateBy = req.UpdateBy
		nodeModels = append(nodeModels, nodeModel)
	}

	if startCount > 1 || endCount > 1 {
		return errors.New("开始节点和结束节点最多只能各有一个")
	}
	if approveCount == 0 {
		return errors.New("至少需要一个审批节点")
	}

	sort.Slice(nodeModels, func(i, j int) bool {
		return nodeModels[i].Sort < nodeModels[j].Sort
	})

	return e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("definition_id = ?", definitionID).Delete(&platformModels.WorkflowDefinitionNode{}).Error; err != nil {
			return err
		}
		return tx.Create(&nodeModels).Error
	})
}

func (e *Workflow) Start(c *gin.Context, req *dto.WorkflowInstanceStartReq) (*WorkflowInstanceDetail, error) {
	req.Normalize()

	currentID := user.GetUserId(c)
	currentName := user.GetUserName(c)

	var module platformModels.ModuleRegistry
	if err := e.Orm.Where("module_key = ? AND status = ?", req.ModuleKey, "2").First(&module).Error; err != nil {
		return nil, errors.New("模块未注册或未启用")
	}

	var definition platformModels.WorkflowDefinition
	if err := e.Orm.Where("definition_id = ? AND status = ?", req.DefinitionId, "2").First(&definition).Error; err != nil {
		return nil, errors.New("流程定义不存在或未启用")
	}
	if definition.ModuleKey != req.ModuleKey || definition.BusinessType != req.BusinessType {
		return nil, errors.New("流程定义与模块/业务类型不匹配")
	}

	approveNodes, err := e.getApproveNodes(definition.DefinitionId)
	if err != nil {
		return nil, err
	}
	if len(approveNodes) == 0 {
		return nil, errors.New("流程定义未配置审批节点")
	}

	firstNode := approveNodes[0]
	firstAssigneeID, firstAssigneeName, err := e.resolveApprover(&firstNode)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if req.Remark == "" {
		req.Remark = "发起审批"
	}

	var detail WorkflowInstanceDetail
	err = e.Orm.Transaction(func(tx *gorm.DB) error {
		instance := platformModels.WorkflowInstance{
			DefinitionId:     definition.DefinitionId,
			DefinitionKey:    definition.DefinitionKey,
			DefinitionName:   definition.DefinitionName,
			ModuleKey:        req.ModuleKey,
			BusinessType:     req.BusinessType,
			BusinessId:       req.BusinessId,
			BusinessNo:       req.BusinessNo,
			Title:            req.Title,
			Status:           dto.WorkflowStatusReview,
			CurrentNodeId:    firstNode.NodeId,
			CurrentNodeKey:   firstNode.NodeKey,
			CurrentNodeName:  firstNode.NodeName,
			StarterId:        currentID,
			StarterName:      currentName,
			StartedAt:        now,
			LastAction:       "start",
			LastActionRemark: req.Remark,
		}
		instance.CreateBy = currentID
		instance.UpdateBy = currentID
		if err := tx.Create(&instance).Error; err != nil {
			return err
		}

		task := platformModels.WorkflowTask{
			InstanceId:   instance.InstanceId,
			DefinitionId: instance.DefinitionId,
			NodeId:       firstNode.NodeId,
			NodeKey:      firstNode.NodeKey,
			NodeName:     firstNode.NodeName,
			AssigneeType: firstNode.ApproverType,
			AssigneeId:   firstAssigneeID,
			AssigneeName: firstAssigneeName,
			Status:       dto.WorkflowTaskPending,
			CreatedAt:    now,
		}
		if err := tx.Create(&task).Error; err != nil {
			return err
		}

		binding := platformModels.WorkflowBusinessBinding{
			ModuleKey:        req.ModuleKey,
			BusinessType:     req.BusinessType,
			BusinessId:       req.BusinessId,
			BusinessNo:       req.BusinessNo,
			Title:            req.Title,
			InstanceId:       instance.InstanceId,
			WorkflowStatus:   dto.WorkflowStatusReview,
			BusinessStatus:   dto.WorkflowStatusReview,
			LastAction:       "start",
			LastActionRemark: req.Remark,
		}
		binding.CreateBy = currentID
		binding.UpdateBy = currentID
		if err := tx.Where("module_key = ? AND business_type = ? AND business_id = ?", req.ModuleKey, req.BusinessType, req.BusinessId).
			Delete(&platformModels.WorkflowBusinessBinding{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&binding).Error; err != nil {
			return err
		}

		actionLog := platformModels.WorkflowActionLog{
			InstanceId:   instance.InstanceId,
			TaskId:       task.TaskId,
			Action:       "start",
			FromStatus:   dto.WorkflowStatusDraft,
			ToStatus:     dto.WorkflowStatusReview,
			ToNodeKey:    firstNode.NodeKey,
			ToNodeName:   firstNode.NodeName,
			OperatorId:   currentID,
			OperatorName: currentName,
			Comment:      req.Remark,
			CreatedAt:    now,
		}
		actionLog.CreateBy = currentID
		actionLog.UpdateBy = currentID
		if err := tx.Create(&actionLog).Error; err != nil {
			return err
		}

		detail.Instance = instance
		detail.Binding = binding
		detail.Tasks = []platformModels.WorkflowTask{task}
		detail.Actions = []platformModels.WorkflowActionLog{actionLog}
		detail.Nodes = approveNodes
		return nil
	})
	if err != nil {
		return nil, err
	}
	return e.GetInstanceDetail(detail.Instance.InstanceId)
}

func (e *Workflow) Approve(c *gin.Context, taskID int, comment string) (*WorkflowInstanceDetail, error) {
	return e.processTask(c, taskID, dto.WorkflowActionApprove, strings.TrimSpace(comment))
}

func (e *Workflow) Reject(c *gin.Context, taskID int, comment string) (*WorkflowInstanceDetail, error) {
	return e.processTask(c, taskID, dto.WorkflowActionReject, strings.TrimSpace(comment))
}

func (e *Workflow) Withdraw(c *gin.Context, instanceID int, comment string) (*WorkflowInstanceDetail, error) {
	comment = strings.TrimSpace(comment)
	currentUserID := user.GetUserId(c)
	currentUserName := user.GetUserName(c)

	var instance platformModels.WorkflowInstance
	if err := e.Orm.First(&instance, instanceID).Error; err != nil {
		return nil, err
	}
	if instance.StarterId != currentUserID {
		return nil, errors.New("只有发起人可以撤回")
	}
	if instance.Status != dto.WorkflowStatusReview {
		return nil, errors.New("当前流程状态不允许撤回")
	}
	if comment == "" {
		comment = "发起人撤回"
	}

	fromNodeKey := instance.CurrentNodeKey
	fromNodeName := instance.CurrentNodeName
	now := time.Now()
	err := e.Orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&platformModels.WorkflowTask{}).
			Where("instance_id = ? AND status = ?", instanceID, dto.WorkflowTaskPending).
			Updates(map[string]interface{}{
				"status":           dto.WorkflowTaskCancelled,
				"action":           dto.WorkflowActionWithdraw,
				"comment":          comment,
				"action_by":        currentUserID,
				"action_by_name":   currentUserName,
				"processed_at":     now,
				"cancelled_reason": comment,
			}).Error; err != nil {
			return err
		}

		instance.Status = dto.WorkflowStatusCanceled
		instance.CurrentNodeId = 0
		instance.CurrentNodeKey = ""
		instance.CurrentNodeName = ""
		instance.LastAction = dto.WorkflowActionWithdraw
		instance.LastActionRemark = comment
		instance.UpdateBy = currentUserID
		instance.FinishedAt = &now
		if err := tx.Save(&instance).Error; err != nil {
			return err
		}

		if err := e.updateBusinessBinding(tx, &instance, dto.WorkflowStatusCanceled, dto.WorkflowStatusCanceled, dto.WorkflowActionWithdraw, comment, currentUserID); err != nil {
			return err
		}
		actionLog := platformModels.WorkflowActionLog{
			InstanceId:   instance.InstanceId,
			Action:       dto.WorkflowActionWithdraw,
			FromStatus:   dto.WorkflowStatusReview,
			ToStatus:     dto.WorkflowStatusCanceled,
			FromNodeKey:  fromNodeKey,
			FromNodeName: fromNodeName,
			OperatorId:   currentUserID,
			OperatorName: currentUserName,
			Comment:      comment,
			CreatedAt:    now,
		}
		actionLog.CreateBy = currentUserID
		actionLog.UpdateBy = currentUserID
		return tx.Create(&actionLog).Error
	})
	if err != nil {
		return nil, err
	}
	return e.GetInstanceDetail(instanceID)
}

func (e *Workflow) GetInstanceDetail(instanceID int) (*WorkflowInstanceDetail, error) {
	var instance platformModels.WorkflowInstance
	if err := e.Orm.First(&instance, instanceID).Error; err != nil {
		return nil, err
	}

	var binding platformModels.WorkflowBusinessBinding
	if err := e.Orm.Where("instance_id = ?", instanceID).First(&binding).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	tasks := make([]platformModels.WorkflowTask, 0)
	if err := e.Orm.Where("instance_id = ?", instanceID).Order("task_id ASC").Find(&tasks).Error; err != nil {
		return nil, err
	}

	actions := make([]platformModels.WorkflowActionLog, 0)
	if err := e.Orm.Where("instance_id = ?", instanceID).Order("log_id ASC").Find(&actions).Error; err != nil {
		return nil, err
	}

	nodes := make([]platformModels.WorkflowDefinitionNode, 0)
	if err := e.Orm.Where("definition_id = ?", instance.DefinitionId).Order("sort ASC, node_id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}

	return &WorkflowInstanceDetail{
		Instance: instance,
		Binding:  binding,
		Tasks:    tasks,
		Actions:  actions,
		Nodes:    nodes,
	}, nil
}

func (e *Workflow) processTask(c *gin.Context, taskID int, action string, comment string) (*WorkflowInstanceDetail, error) {
	currentUserID := user.GetUserId(c)
	currentUserName := user.GetUserName(c)
	currentRoleIDs := authctx.GetRoleIDs(c)

	var task platformModels.WorkflowTask
	if err := e.Orm.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	if task.Status != dto.WorkflowTaskPending {
		return nil, errors.New("当前任务不可处理")
	}
	if !canProcessTask(task, currentUserID, currentRoleIDs) {
		return nil, errors.New("当前用户无权处理该任务")
	}

	var instance platformModels.WorkflowInstance
	if err := e.Orm.First(&instance, task.InstanceId).Error; err != nil {
		return nil, err
	}
	if instance.Status != dto.WorkflowStatusReview {
		return nil, errors.New("当前流程状态不允许处理")
	}

	now := time.Now()
	if comment == "" {
		if action == dto.WorkflowActionApprove {
			comment = "审批通过"
		} else {
			comment = "审批驳回"
		}
	}

	fromNodeKey := instance.CurrentNodeKey
	fromNodeName := instance.CurrentNodeName
	var nextNode *platformModels.WorkflowDefinitionNode
	err := e.Orm.Transaction(func(tx *gorm.DB) error {
		task.Action = action
		task.Comment = comment
		task.ActionBy = currentUserID
		task.ActionByName = currentUserName
		task.ProcessedAt = &now
		if action == dto.WorkflowActionApprove {
			task.Status = dto.WorkflowTaskApproved
		} else {
			task.Status = dto.WorkflowTaskRejected
		}
		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		if action == dto.WorkflowActionApprove {
			nodes, err := e.getApproveNodes(instance.DefinitionId)
			if err != nil {
				return err
			}
			for i := range nodes {
				if nodes[i].NodeId == task.NodeId && i+1 < len(nodes) {
					nextNode = &nodes[i+1]
					break
				}
			}

			if nextNode == nil {
				instance.Status = dto.WorkflowStatusApproved
				instance.CurrentNodeId = 0
				instance.CurrentNodeKey = ""
				instance.CurrentNodeName = ""
				instance.FinishedAt = &now
				instance.LastAction = action
				instance.LastActionRemark = comment
				instance.UpdateBy = currentUserID
				if err := tx.Save(&instance).Error; err != nil {
					return err
				}
				if err := e.updateBusinessBinding(tx, &instance, dto.WorkflowStatusApproved, dto.WorkflowStatusApproved, action, comment, currentUserID); err != nil {
					return err
				}
				actionLog := platformModels.WorkflowActionLog{
					InstanceId:   instance.InstanceId,
					TaskId:       task.TaskId,
					Action:       action,
					FromStatus:   dto.WorkflowStatusReview,
					ToStatus:     dto.WorkflowStatusApproved,
					FromNodeKey:  fromNodeKey,
					FromNodeName: fromNodeName,
					OperatorId:   currentUserID,
					OperatorName: currentUserName,
					Comment:      comment,
					CreatedAt:    now,
				}
				actionLog.CreateBy = currentUserID
				actionLog.UpdateBy = currentUserID
				return tx.Create(&actionLog).Error
			}

			nextAssigneeID, nextAssigneeName, err := e.resolveApprover(nextNode)
			if err != nil {
				return err
			}
			nextTask := platformModels.WorkflowTask{
				InstanceId:   instance.InstanceId,
				DefinitionId: instance.DefinitionId,
				NodeId:       nextNode.NodeId,
				NodeKey:      nextNode.NodeKey,
				NodeName:     nextNode.NodeName,
				AssigneeType: nextNode.ApproverType,
				AssigneeId:   nextAssigneeID,
				AssigneeName: nextAssigneeName,
				Status:       dto.WorkflowTaskPending,
				CreatedAt:    now,
			}
			if err := tx.Create(&nextTask).Error; err != nil {
				return err
			}

			instance.CurrentNodeId = nextNode.NodeId
			instance.CurrentNodeKey = nextNode.NodeKey
			instance.CurrentNodeName = nextNode.NodeName
			instance.LastAction = action
			instance.LastActionRemark = comment
			instance.UpdateBy = currentUserID
			if err := tx.Save(&instance).Error; err != nil {
				return err
			}
			if err := e.updateBusinessBinding(tx, &instance, dto.WorkflowStatusReview, dto.WorkflowStatusReview, action, comment, currentUserID); err != nil {
				return err
			}
			actionLog := platformModels.WorkflowActionLog{
				InstanceId:   instance.InstanceId,
				TaskId:       task.TaskId,
				Action:       action,
				FromStatus:   dto.WorkflowStatusReview,
				ToStatus:     dto.WorkflowStatusReview,
				FromNodeKey:  fromNodeKey,
				FromNodeName: fromNodeName,
				ToNodeKey:    nextNode.NodeKey,
				ToNodeName:   nextNode.NodeName,
				OperatorId:   currentUserID,
				OperatorName: currentUserName,
				Comment:      comment,
				CreatedAt:    now,
			}
			actionLog.CreateBy = currentUserID
			actionLog.UpdateBy = currentUserID
			return tx.Create(&actionLog).Error
		}

		instance.Status = dto.WorkflowStatusRejected
		instance.LastAction = action
		instance.LastActionRemark = comment
		instance.UpdateBy = currentUserID
		instance.FinishedAt = &now
		if err := tx.Save(&instance).Error; err != nil {
			return err
		}
		if err := e.updateBusinessBinding(tx, &instance, dto.WorkflowStatusRejected, dto.WorkflowStatusRejected, action, comment, currentUserID); err != nil {
			return err
		}
		actionLog := platformModels.WorkflowActionLog{
			InstanceId:   instance.InstanceId,
			TaskId:       task.TaskId,
			Action:       action,
			FromStatus:   dto.WorkflowStatusReview,
			ToStatus:     dto.WorkflowStatusRejected,
			FromNodeKey:  fromNodeKey,
			FromNodeName: fromNodeName,
			OperatorId:   currentUserID,
			OperatorName: currentUserName,
			Comment:      comment,
			CreatedAt:    now,
		}
		actionLog.CreateBy = currentUserID
		actionLog.UpdateBy = currentUserID
		return tx.Create(&actionLog).Error
	})
	if err != nil {
		return nil, err
	}
	return e.GetInstanceDetail(instance.InstanceId)
}

func (e *Workflow) getApproveNodes(definitionID int) ([]platformModels.WorkflowDefinitionNode, error) {
	nodes := make([]platformModels.WorkflowDefinitionNode, 0)
	if err := e.Orm.Where("definition_id = ?", definitionID).Order("sort ASC, node_id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	result := make([]platformModels.WorkflowDefinitionNode, 0)
	for _, node := range nodes {
		if node.NodeType == dto.WorkflowNodeTypeApprove {
			result = append(result, node)
		}
	}
	return result, nil
}

func (e *Workflow) resolveApprover(node *platformModels.WorkflowDefinitionNode) (int, string, error) {
	switch node.ApproverType {
	case dto.WorkflowApproverUser:
		userID, err := strconv.Atoi(node.ApproverValue)
		if err != nil {
			return 0, "", errors.New("审批用户配置非法")
		}
		var model adminModels.SysUser
		if err := e.Orm.First(&model, userID).Error; err != nil {
			return 0, "", errors.New("审批用户不存在")
		}
		name := node.ApproverName
		if name == "" {
			name = model.NickName
			if name == "" {
				name = model.Username
			}
		}
		return userID, name, nil
	case dto.WorkflowApproverRole:
		roleID, err := strconv.Atoi(node.ApproverValue)
		if err != nil {
			return 0, "", errors.New("审批角色配置非法")
		}
		var model adminModels.SysRole
		if err := e.Orm.First(&model, roleID).Error; err != nil {
			return 0, "", errors.New("审批角色不存在")
		}
		name := node.ApproverName
		if name == "" {
			name = model.RoleName
		}
		return roleID, name, nil
	default:
		return 0, "", errors.New("不支持的审批人类型")
	}
}

func (e *Workflow) updateBusinessBinding(tx *gorm.DB, instance *platformModels.WorkflowInstance, workflowStatus, businessStatus, action, comment string, currentUserID int) error {
	return tx.Model(&platformModels.WorkflowBusinessBinding{}).
		Where("instance_id = ?", instance.InstanceId).
		Updates(map[string]interface{}{
			"workflow_status":    workflowStatus,
			"business_status":    businessStatus,
			"last_action":        action,
			"last_action_remark": comment,
			"update_by":          currentUserID,
		}).Error
}

func canProcessTask(task platformModels.WorkflowTask, currentUserID int, currentRoleIDs []int) bool {
	switch task.AssigneeType {
	case dto.WorkflowApproverUser:
		return task.AssigneeId == currentUserID
	case dto.WorkflowApproverRole:
		return slices.Contains(currentRoleIDs, task.AssigneeId)
	default:
		return false
	}
}
