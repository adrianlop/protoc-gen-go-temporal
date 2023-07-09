package plugin

import "github.com/iancoleman/strcase"

type Names struct {
	*Service
}

func (svc *Service) Names() *Names {
	return &Names{svc}
}

func (svc *Names) ActivityGoName(activity string) string {
	return strcase.ToCamel(activity)
}

func (svc *Names) ActivitiesInterface() string {
	return strcase.ToCamel(svc.GoName + "Activities")
}

func (svc *Names) ActivityFutureImpl(activity string) string {
	return strcase.ToLowerCamel(svc.GoName + activity + "Future")
}

func (svc *Names) ActivityName(activity string) string {
	if n := svc.activities[activity].GetName(); n != "" {
		return n
	}
	return string(svc.methods[activity].Desc.FullName())
}

func (svc *Names) ActivityNameConstant(activity string) string {
	return strcase.ToCamel(svc.GoName + activity + "ActivityName")
}

func (svc *Names) ActivityOptions(activity string) string {
	return strcase.ToCamel(svc.GoName + activity + "Options")
}

func (svc *Names) Builder(workflow string) string {
	return strcase.ToLowerCamel("build" + svc.GoName + workflow)
}

func (svc *Names) ClientImpl() string {
	return strcase.ToLowerCamel(svc.GoName + "Client")
}

func (svc *Names) ClientInterface() string {
	return strcase.ToCamel(svc.GoName + "Client")
}

func (svc *Names) IDExpression(method string) string {
	return strcase.ToCamel(svc.GoName + method + "IdExpression")
}

func (svc *Names) QueryGoName(query string) string {
	return strcase.ToCamel(query)
}

func (svc *Names) QueryName(query string) string {
	if n := svc.queries[query].GetName(); n != "" {
		return n
	}
	return string(svc.methods[query].Desc.FullName())
}

func (svc *Names) QueryNameConstant(query string) string {
	return strcase.ToCamel(svc.GoName + query + "QueryName")
}

func (svc *Names) RegisterActivities() string {
	return strcase.ToCamel("register" + svc.GoName + "Activities")
}

func (svc *Names) RegisterActivitiy(activity string) string {
	return strcase.ToCamel("register" + svc.GoName + activity + "Activity")
}

func (svc *Names) RegisterWorkflow(workflow string) string {
	return strcase.ToCamel("register" + svc.GoName + workflow + "Workflow")
}

func (svc *Names) RegisterWorkflows() string {
	return strcase.ToCamel("register" + svc.GoName + "Workflows")
}

func (svc *Names) Self() string {
	return strcase.ToCamel(svc.GoName)
}

func (svc *Names) SignalGoName(signal string) string {
	return strcase.ToCamel(signal)
}

func (svc *Names) SignalImpl(signal string) string {
	return strcase.ToCamel(svc.GoName + signal + "Signal")
}

func (svc *Names) SignalName(signal string) string {
	if n := svc.signals[signal].GetName(); n != "" {
		return n
	}
	return string(svc.methods[signal].Desc.FullName())
}

func (svc *Names) SignalNameConstant(signal string) string {
	return strcase.ToCamel(svc.GoName + signal + "SignalName")
}

func (svc *Names) TaskQueue() string {
	return strcase.ToCamel(svc.GoName + "TaskQueue")
}

func (svc *Names) UpdateGoName(udpate string) string {
	return strcase.ToCamel(udpate)
}

func (svc *Names) UpdateHandleImpl(update string) string {
	return strcase.ToLowerCamel(svc.GoName + update + "Handle")
}

func (svc *Names) UpdateHandleInterface(update string) string {
	return strcase.ToCamel(svc.GoName + update + "Handle")
}

func (svc *Names) UpdateName(update string) string {
	if n := svc.updates[update].GetName(); n != "" {
		return n
	}
	return string(svc.methods[update].Desc.FullName())
}

func (svc *Names) UpdateNameConstant(update string) string {
	return strcase.ToCamel(svc.GoName + update + "UpdateName")
}

func (svc *Names) WorkflowChildRunImpl(workflow string) string {
	return strcase.ToLowerCamel(svc.GoName + workflow + "ChildRun")
}

func (svc *Names) WorkflowChildRunInterface(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow + "ChildRun")
}

func (svc *Names) WorkflowGoName(workflow string) string {
	return strcase.ToCamel(workflow)
}

func (svc *Names) WorkflowImpl(workflow string) string {
	return strcase.ToLowerCamel(svc.GoName + workflow)
}

func (svc *Names) WorkflowInput(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow + "Input")
}

func (svc *Names) WorkflowInterface(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow)
}

func (svc *Names) WorkflowName(workflow string) string {
	if n := svc.workflows[workflow].GetName(); n != "" {
		return n
	}
	return string(svc.methods[workflow].Desc.FullName())
}

func (svc *Names) WorkflowNameConstant(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow + "WorkflowName")
}

func (svc *Names) WorkflowOptions(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow + "Options")
}

func (svc *Names) WorkflowResources() string {
	return strcase.ToCamel(svc.GoName + "WorkflowResources")
}

func (svc *Names) WorkflowRunImpl(workflow string) string {
	return strcase.ToLowerCamel(svc.GoName + workflow + "Run")
}

func (svc *Names) WorkflowRunInterface(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow + "Run")
}

func (svc *Names) WorkflowSearchAttributesMapping(workflow string) string {
	return strcase.ToCamel(svc.GoName + workflow + "SearchAttributesMapping")
}

func (svc *Names) WorkflowsInterface() string {
	return strcase.ToCamel(svc.GoName + "Workflows")
}
