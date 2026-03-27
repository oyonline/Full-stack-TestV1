package jobs

import (
	"fmt"
	"go-admin/app/admin/service"
	"go-admin/app/admin/service/dto"
	"go-admin/common/utils"
	"time"

	"github.com/go-admin-team/go-admin-core/logger"
)

// InitJob
// 需要将定义的struct 添加到字典中；
// 字典 key 可以配置到 自动任务 调用目标 中；
func InitJob() {
	jobList = map[string]JobExec{
		"ExamplesOne":        ExamplesOne{},
		"CostCenterTimeTask": CostCenterTimeTask{},
		"PullDeptTask":       PullDeptTask{},
		"PullCustomerTask":   PullCustomerTask{},
		// ...
	}
}

// ExamplesOne
// 新添加的job 必须按照以下格式定义，并实现Exec函数
type ExamplesOne struct {
}

func (t ExamplesOne) Exec(arg interface{}) error {
	str := time.Now().Format(timeFormat) + " [INFO] JobCore ExamplesOne exec success"
	// TODO: 这里需要注意 Examples 传入参数是 string 所以 arg.(string)；请根据对应的类型进行转化；
	switch arg.(type) {

	case string:
		if arg.(string) != "" {
			fmt.Println("string", arg.(string))
			fmt.Println(str, arg.(string))
		} else {
			fmt.Println("arg is nil")
			fmt.Println(str, "arg is nil")
		}
		break
	}

	return nil
}

type CostCenterTimeTask struct {
}

func (t CostCenterTimeTask) Exec(arg interface{}) error {
	startTime := time.Now()
	logPrefix := startTime.Format(timeFormat) + " [CostCenterTimeTask]"
	fmt.Println(logPrefix, "=== 任务开始执行 ===")
	// 2. 创建 Service 实例
	costCenterService := &service.CostCenterInfo{}
	costCenterService.Orm = utils.GetDb()
	costCenterService.Log = logger.NewHelper(logger.DefaultLogger)

	err := costCenterService.CostCenterTimeTask()
	execTime := time.Since(startTime)
	if err != nil {
		fmt.Printf("%s [ERROR] 执行失败：%v, 耗时：%v\n", logPrefix, err, execTime)
		return err
	}
	fmt.Printf("%s [SUCCESS] 成功启用, 耗时：%v\n", logPrefix, execTime)
	return nil
}

// PullDeptTask
type PullDeptTask struct {
}

func (t PullDeptTask) Exec(arg interface{}) error {
	startTime := time.Now()
	logPrefix := startTime.Format(timeFormat) + " [PullDeptTask]"
	fmt.Println(logPrefix, "开始执行！", arg)

	s := &service.SysDept{}
	s.Orm = utils.GetDb()
	s.Log = logger.NewHelper(logger.DefaultLogger)

	s.Su = &service.SysUser{}
	s.Su.Orm = utils.GetDb()
	s.Su.Log = logger.NewHelper(logger.DefaultLogger)

	req := dto.DepartmentBatch{}
	req.OpenDepartmentIds = []string{"0"}
	req.ParentId = 0
	req.SetCreateBy(1)

	err := s.PullDepartmentChildrens(&req)
	execTime := time.Since(startTime)
	if err != nil {
		fmt.Printf("%s [ERROR] 执行失败：%v, 耗时：%v\n", logPrefix, err, execTime)
		return err
	}
	fmt.Printf("%s [SUCCESS] 执行成功, 耗时：%v\n", logPrefix, execTime)
	return nil
}

// PullCustomerTask
type PullCustomerTask struct {
}

func (t PullCustomerTask) Exec(arg interface{}) error {
	startTime := time.Now()
	logPrefix := startTime.Format(timeFormat) + " [PullCustomerTask]"
	fmt.Println(logPrefix, "开始执行！", arg)

	s := &service.KingdeeCustomer{}
	s.Orm = utils.GetDb()
	s.Log = logger.NewHelper(logger.DefaultLogger)
	createBy := 1

	err := s.PullKingdeeCustomers(&createBy)
	execTime := time.Since(startTime)
	if err != nil {
		fmt.Printf("%s [ERROR] 执行失败：%v, 耗时：%v\n", logPrefix, err, execTime)
		return err
	}
	fmt.Printf("%s [SUCCESS] 执行成功, 耗时：%v\n", logPrefix, execTime)
	return nil
}
