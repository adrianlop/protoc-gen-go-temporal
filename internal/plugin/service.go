package plugin

import (
	"errors"
	"fmt"
	"sort"

	temporalv1 "github.com/cludden/protoc-gen-go-temporal/gen/temporal/v1"
	g "github.com/dave/jennifer/jen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

// imported packages
const (
	activityPkg   = "go.temporal.io/sdk/activity"
	clientPkg     = "go.temporal.io/sdk/client"
	enumsPkg      = "go.temporal.io/api/enums/v1"
	expressionPkg = "github.com/cludden/protoc-gen-go-temporal/pkg/expression"
	temporalPkg   = "go.temporal.io/sdk/temporal"
	updatePkg     = "go.temporal.io/api/update/v1"
	uuidPkg       = "github.com/google/uuid"
	workflowPkg   = "go.temporal.io/sdk/workflow"
	workerPkg     = "go.temporal.io/sdk/worker"
)

// method modes
const (
	modeActivity uint8 = 1 << iota
	modeQuery
	modeSignal
	modeUpdate
	modeWorkflow
)

// Service describes a temporal protobuf service definition
type Service struct {
	*protogen.Plugin
	*protogen.Service
	*protogen.File
	opts              *temporalv1.ServiceOptions
	activitiesOrdered []string
	activities        map[string]*temporalv1.ActivityOptions
	methods           map[string]*protogen.Method
	queriesOrdered    []string
	queries           map[string]*temporalv1.QueryOptions
	signalsOrdered    []string
	signals           map[string]*temporalv1.SignalOptions
	updatesOrdered    []string
	updates           map[string]*temporalv1.UpdateOptions
	workflowsOrdered  []string
	workflows         map[string]*temporalv1.WorkflowOptions
}

// parseService extracts a Service from a protogen.Service value
func parseService(p *protogen.Plugin, file *protogen.File, service *protogen.Service) (*Service, error) {
	svc := Service{
		Plugin:     p,
		Service:    service,
		File:       file,
		activities: make(map[string]*temporalv1.ActivityOptions),
		methods:    make(map[string]*protogen.Method),
		queries:    make(map[string]*temporalv1.QueryOptions),
		signals:    make(map[string]*temporalv1.SignalOptions),
		updates:    make(map[string]*temporalv1.UpdateOptions),
		workflows:  make(map[string]*temporalv1.WorkflowOptions),
	}

	if opts, ok := proto.GetExtension(service.Desc.Options(), temporalv1.E_Service).(*temporalv1.ServiceOptions); ok && opts != nil {
		svc.opts = opts
	}

	for _, method := range service.Methods {
		name := method.GoName
		svc.methods[name] = method
		var mode uint8

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions); ok && opts != nil {
			svc.activities[name] = opts
			svc.activitiesOrdered = append(svc.activitiesOrdered, name)
			mode |= modeActivity
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions); ok && opts != nil {
			svc.workflows[name] = opts
			svc.workflowsOrdered = append(svc.workflowsOrdered, name)
			mode |= modeWorkflow
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions); ok && opts != nil {
			if mode > 0 {
				return nil, fmt.Errorf("error parsing %q method: query is incompatible with other method options", method.Desc.FullName())
			}
			svc.queries[name] = opts
			svc.queriesOrdered = append(svc.queriesOrdered, name)
			mode |= modeQuery
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions); ok && opts != nil {
			if mode > 0 {
				return nil, fmt.Errorf("error parsing %q method: signal is incompatible with other method options", method.Desc.FullName())
			}
			svc.signals[name] = opts
			svc.signalsOrdered = append(svc.signalsOrdered, name)
			mode |= modeSignal
		}

		if svc.opts.GetFeatures().GetWorkflowUpdate().GetEnabled() {
			if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Update).(*temporalv1.UpdateOptions); ok && opts != nil {
				if mode > 0 {
					return nil, fmt.Errorf("error parsing %q method: update is incompatible with other method options", method.Desc.FullName())
				}
				svc.updates[name] = opts
				svc.updatesOrdered = append(svc.updatesOrdered, name)
				mode |= modeUpdate
			}
		}
	}
	sort.Strings(svc.activitiesOrdered)
	sort.Strings(svc.queriesOrdered)
	sort.Strings(svc.signalsOrdered)
	sort.Strings(svc.updatesOrdered)
	sort.Strings(svc.workflowsOrdered)

	var errs error
	for _, workflow := range svc.workflowsOrdered {
		opts := svc.workflows[workflow]

		// ensure workflow queries are defined
		for _, queryOpts := range opts.GetQuery() {
			query := queryOpts.GetRef()
			if _, ok := svc.queries[query]; !ok {
				errs = errors.Join(errs, fmt.Errorf("workflow  %q references undefined query: %q", workflow, query))
			}
		}

		// ensure workflow signals are defined
		for _, signalOpts := range opts.GetSignal() {
			signal := signalOpts.GetRef()
			if _, ok := svc.signals[signal]; !ok {
				errs = errors.Join(errs, fmt.Errorf("workflow  %q references undefined signal: %q", workflow, signal))
			}
		}

		// ensure workflow updates are defined
		for _, updateOpts := range opts.GetUpdate() {
			update := updateOpts.GetRef()
			if _, ok := svc.updates[update]; !ok {
				errs = errors.Join(errs, fmt.Errorf("workflow  %q references undefined update: %q", workflow, update))
			}
		}
	}

	// ensure that signals return no value, unless signal method is also an activity, query, and/or workflow
	for _, signal := range svc.signalsOrdered {
		handler := svc.methods[signal]
		if !isEmpty(handler.Output) {
			errs = errors.Join(errs, fmt.Errorf("expected signal %q output to be google.protobuf.Empty, got: %s", signal, handler.Output.GoIdent.GoName))
		}
	}
	return &svc, errs
}

// genConstants generates constants
func (svc *Service) genConstants(f *g.File) {
	// add task queue
	if taskQueue := svc.opts.GetTaskQueue(); taskQueue != "" {
		f.Commentf("%s is the default task-queue for a %s worker", svc.Names().TaskQueue(), svc.GoName)
		f.Const().Id(svc.Names().TaskQueue()).Op("=").Lit(taskQueue)
	}

	// add workflow names
	if len(svc.workflows) > 0 {
		f.Commentf("%s workflow names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, workflow := range svc.workflowsOrdered {
				defs.Id(svc.Names().WorkflowNameConstant(workflow)).Op("=").Lit(svc.Names().WorkflowName(workflow))
			}
		})
	}

	// add workflow id expressions
	workflowIdExpressions := [][]string{}
	for _, workflow := range svc.workflowsOrdered {
		opts := svc.workflows[workflow]
		if expr := opts.GetId(); expr != "" {
			workflowIdExpressions = append(workflowIdExpressions, []string{workflow, expr})
		}
	}
	if len(workflowIdExpressions) > 0 {
		f.Commentf("%s workflow id expressions", svc.GoName)
		f.Var().DefsFunc(func(defs *g.Group) {
			for _, pair := range workflowIdExpressions {
				defs.Id(svc.Names().IDExpression(pair[0])).Op("=").Qual(expressionPkg, "MustParseExpression").Call(g.Lit(pair[1]))
			}
		})
	}

	// add workflow search attribute mappings
	workflowSearchAttributes := [][]string{}
	for _, workflow := range svc.workflowsOrdered {
		opts := svc.workflows[workflow]
		if mapping := opts.GetSearchAttributes(); mapping != "" {
			workflowSearchAttributes = append(workflowSearchAttributes, []string{workflow, mapping})
		}
	}
	if len(workflowSearchAttributes) > 0 {
		f.Commentf("%s workflow search attribute mappings", svc.GoName)
		f.Var().DefsFunc(func(defs *g.Group) {
			for _, pair := range workflowSearchAttributes {
				defs.Id(svc.Names().WorkflowSearchAttributesMapping(pair[0])).Op("=").Qual(expressionPkg, "MustParseMapping").Call(g.Lit(pair[1]))
			}
		})
	}

	// add activity names
	if len(svc.activities) > 0 {
		f.Commentf("%s activity names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, activity := range svc.activitiesOrdered {
				defs.Id(svc.Names().ActivityNameConstant(activity)).Op("=").Lit(svc.Names().ActivityName(activity))
			}
		})
	}

	// add query names
	if len(svc.queries) > 0 {
		f.Commentf("%s query names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, query := range svc.queriesOrdered {
				defs.Id(svc.Names().QueryNameConstant(query)).Op("=").Lit(svc.Names().QueryName(query))
			}
		})
	}

	// add signal names
	if len(svc.signals) > 0 {
		f.Commentf("%s signal names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, signal := range svc.signalsOrdered {
				defs.Id(svc.Names().UpdateNameConstant(signal)).Op("=").Lit(svc.Names().SignalName(signal))
			}
		})
	}

	// add update names
	if len(svc.updates) > 0 {
		f.Commentf("%s update names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, update := range svc.updatesOrdered {
				defs.Id(svc.Names().UpdateNameConstant(update)).Op("=").Lit(svc.Names().UpdateName(update))
			}
		})
	}

	// add update id expressions
	updateIdExpressions := [][]string{}
	for _, update := range svc.updatesOrdered {
		opts := svc.updates[update]
		if expr := opts.GetId(); expr != "" {
			updateIdExpressions = append(updateIdExpressions, []string{update, expr})
		}
	}
	if len(updateIdExpressions) > 0 {
		f.Commentf("%s update id expressions", svc.GoName)
		f.Var().DefsFunc(func(defs *g.Group) {
			for _, pair := range updateIdExpressions {
				defs.Id(svc.Names().IDExpression(pair[0])).Op("=").Qual(expressionPkg, "MustParseExpression").Call(g.Lit(pair[1]))
			}
		})
	}
}

// render writes the temporal service to the given File
func (svc *Service) render(f *g.File) {
	svc.genConstants(f)

	// generate client interface and implementation
	svc.genClientInterface(f)
	svc.genClientImpl(f)
	svc.genClientImplConstructor(f)

	// generate client workflow methods
	for _, workflow := range svc.workflowsOrdered {
		opts := svc.workflows[workflow]
		svc.genClientImplWorkflowMethod(f, workflow)
		svc.genClientImplWorkflowAsyncMethod(f, workflow)
		svc.genClientImplWorkflowGetMethod(f, workflow)
		for _, signal := range opts.GetSignal() {
			if signal.GetStart() {
				svc.genClientImplSignalWithStartMethod(f, workflow, signal.GetRef())
				svc.genClientImplSignalWithStartAsyncMethod(f, workflow, signal.GetRef())
			}
		}
	}

	// generate client query methods
	for _, query := range svc.queriesOrdered {
		svc.genClientImplQueryMethod(f, query)
	}

	// generate client signal methods
	for _, signal := range svc.signalsOrdered {
		svc.genClientImplSignalMethod(f, signal)
	}

	// generate client update methods
	for _, update := range svc.updatesOrdered {
		svc.genClientImplUpdateMethod(f, update)
		svc.genClientImplUpdateMethodAsync(f, update)
	}

	// generate <Workflow>Run interfaces and implementations used by client
	for _, workflow := range svc.workflowsOrdered {
		opts := svc.workflows[workflow]
		svc.genClientWorkflowRunInterface(f, workflow)
		svc.genClientWorkflowRunImpl(f, workflow)
		svc.genClientWorkflowRunImplIDMethod(f, workflow)
		svc.genClientWorkflowRunImplRunIDMethod(f, workflow)
		svc.genClientWorkflowRunImplGetMethod(f, workflow)

		// generate query methods
		for _, queryOpts := range opts.GetQuery() {
			svc.genClientWorkflowRunImplQueryMethod(f, workflow, queryOpts.GetRef())
		}

		// generate signal methods
		for _, signalOpts := range opts.GetSignal() {
			svc.genClientWorkflowRunImplSignalMethod(f, workflow, signalOpts.GetRef())
		}

		// generate update methods
		for _, updateOpts := range opts.GetUpdate() {
			svc.genClientWorkflowRunImplUpdateMethod(f, workflow, updateOpts.GetRef())
			svc.genClientWorkflowRunImplUpdateAsyncMethod(f, workflow, updateOpts.GetRef())
		}
	}

	// generate <Update>Handle interfaces and implementations used by client
	for _, update := range svc.updatesOrdered {
		svc.genClientUpdateHandleInterface(f, update)
		svc.genClientUpdateHandleImpl(f, update)
		svc.genClientUpdateHandleImplWorkflowIDMethod(f, update)
		svc.genClientUpdateHandleImplRunIDMethod(f, update)
		svc.genClientUpdateHandleImplUpdateIDMethod(f, update)
		svc.genClientUpdateHandleImplGetMethod(f, update)
	}

	// generate workflows interface and registration helper
	svc.genWorkerWorkflowsInterface(f)
	svc.genWorkerRegisterWorkflows(f)

	// generate workflow resources
	svc.genWorkerWorkflowResources(f)
	for _, activity := range svc.activitiesOrdered {
		svc.genWorkerWorkflowResourcesActivityMethod(f, activity, false, false)
		svc.genWorkerWorkflowResourcesActivityMethod(f, activity, false, true)
		svc.genWorkerWorkflowResourcesActivityMethod(f, activity, true, false)
		svc.genWorkerWorkflowResourcesActivityMethod(f, activity, true, true)
	}
	for _, signal := range svc.signalsOrdered {
		svc.genWorkerWorkflowResourcesSignalExternal(f, signal)
	}
	for _, workflow := range svc.workflowsOrdered {
		svc.genWorkerWorkflowResourcesChildWorkflowMethod(f, workflow)
		svc.genWorkerWorkflowResourcesChildWorkflowAsyncMethod(f, workflow)
	}

	// generate workflow types, methods, functions
	for _, workflow := range svc.workflowsOrdered {
		svc.genWorkerRegisterWorkflow(f, workflow)
		svc.genWorkerWorkflowBuilder(f, workflow)
		//svc.genWorkerWorkflowFactory(f, workflow)
		//svc.genWorkerWorkflowFactoryExecuteMethod(f, workflow)
		svc.genWorkerWorkflowInput(f, workflow)
		svc.genWorkerWorkflowInterface(f, workflow)
		svc.genWorkerWorkflowResourcesChildWorkflowMethod(f, workflow)
		svc.genWorkerWorkflowResourcesChildWorkflowAsyncMethod(f, workflow)
		svc.genWorkerWorkflowChildRunImpl(f, workflow)
		svc.genWorkerWorkflowChildRunImplGetMethod(f, workflow)
		svc.genWorkerWorkflowChildRunImplSelectMethod(f, workflow)
		svc.genWorkerWorkflowChildRunImplSelectStartMethod(f, workflow)
		svc.genWorkerWorkflowChildRunIpmlWaitStartMethod(f, workflow)
		svc.genWorkerWorkflowChildRunImplSignalMethods(f, workflow)
	}

	// generate signal types, methods, functions
	for _, signal := range svc.signalsOrdered {
		svc.genWorkerSignal(f, signal)
		svc.genWorkerSignalReceive(f, signal)
		svc.genWorkerSignalReceiveAsync(f, signal)
		svc.genWorkerSignalSelect(f, signal)
		svc.genWorkerWorkflowResourcesSignalExternal(f, signal)
	}

	// generate activities
	svc.genActivitiesInterface(f)
	svc.genRegisterActivities(f)
	for _, activity := range svc.activitiesOrdered {
		svc.genRegisterActivity(f, activity)
		svc.genActivityFuture(f, activity)
		svc.genActivityFutureGetMethod(f, activity)
		svc.genActivityFutureSelectMethod(f, activity)
	}
}
