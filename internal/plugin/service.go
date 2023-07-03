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
	testsuitePkg  = "go.temporal.io/sdk/testsuite"
	updatePkg     = "go.temporal.io/api/update/v1"
	uuidPkg       = "github.com/google/uuid"
	workflowPkg   = "go.temporal.io/sdk/workflow"
	workerPkg     = "go.temporal.io/sdk/worker"
)

// Service describes a temporal protobuf service definition
type Service struct {
	*protogen.Plugin
	*protogen.Service
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
func parseService(p *protogen.Plugin, service *protogen.Service) (*Service, error) {
	svc := Service{
		Plugin:     p,
		Service:    service,
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

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Activity).(*temporalv1.ActivityOptions); ok && opts != nil {
			svc.activities[name] = opts
			svc.activitiesOrdered = append(svc.activitiesOrdered, name)
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Query).(*temporalv1.QueryOptions); ok && opts != nil {
			svc.queries[name] = opts
			svc.queriesOrdered = append(svc.queriesOrdered, name)
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Signal).(*temporalv1.SignalOptions); ok && opts != nil {
			svc.signals[name] = opts
			svc.signalsOrdered = append(svc.signalsOrdered, name)
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Update).(*temporalv1.UpdateOptions); ok && opts != nil {
			svc.updates[name] = opts
			svc.updatesOrdered = append(svc.updatesOrdered, name)
		}

		if opts, ok := proto.GetExtension(method.Desc.Options(), temporalv1.E_Workflow).(*temporalv1.WorkflowOptions); ok && opts != nil {
			svc.workflows[name] = opts
			svc.workflowsOrdered = append(svc.workflowsOrdered, name)
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
		_, isActivity := svc.activities[signal]
		_, isQuery := svc.queries[signal]
		_, isUpdate := svc.updates[signal]
		_, isWorkflow := svc.workflows[signal]
		if !isActivity && !isQuery && !isUpdate && !isWorkflow && !isEmpty(handler.Output) {
			errs = errors.Join(errs, fmt.Errorf("expected signal %q output to be google.protobuf.Empty, got: %s", signal, handler.Output.GoIdent.GoName))
		}
	}
	return &svc, errs
}

// genConstants generates constants
func (svc *Service) genConstants(f *g.File) {
	// add task queue
	if taskQueue := svc.opts.GetTaskQueue(); taskQueue != "" {
		f.Commentf("%sTaskQueue is the default task-queue for a %s worker", svc.GoName, svc.GoName)
		f.Const().Id(fmt.Sprintf("%sTaskQueue", svc.GoName)).Op("=").Lit(taskQueue)
	}

	// add workflow names
	if len(svc.workflows) > 0 {
		f.Commentf("%s workflow names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, workflow := range svc.workflowsOrdered {
				method := svc.methods[workflow]
				opts := svc.workflows[workflow]
				name := opts.GetName()
				if name == "" {
					name = string(method.Desc.FullName())
				}
				defs.Id(fmt.Sprintf("%sWorkflowName", workflow)).Op("=").Lit(name)
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
				defs.Id(fmt.Sprintf("%sIDExpression", pair[0])).Op("=").Qual(expressionPkg, "MustParseExpression").Call(g.Lit(pair[1]))
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
				defs.Id(fmt.Sprintf("%sSearchAttributesMapping", pair[0])).Op("=").Qual(expressionPkg, "MustParseMapping").Call(g.Lit(pair[1]))
			}
		})
	}

	// add activity names
	if len(svc.activities) > 0 {
		f.Commentf("%s activity names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, activity := range svc.activitiesOrdered {
				method := svc.methods[activity]
				opts := svc.activities[activity]
				name := opts.GetName()
				if name == "" {
					name = string(method.Desc.FullName())
				}
				defs.Id(fmt.Sprintf("%sActivityName", activity)).Op("=").Lit(name)
			}
		})
	}

	// add query names
	if len(svc.queries) > 0 {
		f.Commentf("%s query names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, query := range svc.queriesOrdered {
				method := svc.methods[query]
				opts := svc.queries[query]
				name := opts.GetName()
				if name == "" {
					name = string(method.Desc.FullName())
				}
				defs.Id(fmt.Sprintf("%sQueryName", query)).Op("=").Lit(name)
			}
		})
	}

	// add signal names
	if len(svc.signals) > 0 {
		f.Commentf("%s signal names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, signal := range svc.signalsOrdered {
				method := svc.methods[signal]
				opts := svc.signals[signal]
				name := opts.GetName()
				if name == "" {
					name = string(method.Desc.FullName())
				}
				defs.Id(fmt.Sprintf("%sSignalName", signal)).Op("=").Lit(name)
			}
		})
	}

	// add update names
	if len(svc.updates) > 0 {
		f.Commentf("%s update names", svc.GoName)
		f.Const().DefsFunc(func(defs *g.Group) {
			for _, update := range svc.updatesOrdered {
				method := svc.methods[update]
				opts := svc.updates[update]
				name := opts.GetName()
				if name == "" {
					name = string(method.Desc.FullName())
				}
				defs.Id(fmt.Sprintf("%sUpdateName", update)).Op("=").Lit(name)
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
				defs.Id(fmt.Sprintf("%sIDExpression", pair[0])).Op("=").Qual(expressionPkg, "MustParseExpression").Call(g.Lit(pair[1]))
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

	// generate workflow types, methods, functions
	for _, workflow := range svc.workflowsOrdered {
		svc.genWorkerRegisterWorkflow(f, workflow)
		svc.genWorkerBuilderFunction(f, workflow)
		svc.genWorker(f, workflow)
		svc.genWorkerExecuteMethod(f, workflow)
		svc.genWorkerWorkflowInput(f, workflow)
		svc.genWorkerWorkflowInterface(f, workflow)
		svc.genWorkerChildWorkflow(f, workflow)
		svc.genWorkerChildWorkflowAsync(f, workflow)
		svc.genWorkerWorkflowChildRun(f, workflow)
		svc.genWorkerWorkflowChildRunGet(f, workflow)
		svc.genWorkerWorkflowChildRunSelect(f, workflow)
		svc.genWorkerWorkflowChildRunSelectStart(f, workflow)
		svc.genWorkerWorkflowChildRunWaitStart(f, workflow)
		svc.genWorkerWorkflowChildRunSignals(f, workflow)
	}

	// generate signal types, methods, functions
	for _, signal := range svc.signalsOrdered {
		svc.genWorkerSignal(f, signal)
		svc.genWorkerSignalReceive(f, signal)
		svc.genWorkerSignalReceiveAsync(f, signal)
		svc.genWorkerSignalSelect(f, signal)
		svc.genWorkerSignalExternal(f, signal)
	}

	// generate activities
	svc.genActivitiesInterface(f)
	svc.genRegisterActivities(f)
	for _, activity := range svc.activitiesOrdered {
		svc.genRegisterActivity(f, activity)
		svc.genActivityFuture(f, activity)
		svc.genActivityFutureGetMethod(f, activity)
		svc.genActivityFutureSelectMethod(f, activity)
		svc.genActivityFunction(f, activity, false)
		svc.genActivityFunction(f, activity, true)
	}

	// generate test client
	svc.genTestClientImpl(f)
	svc.genTestClientImplNewMethod(f)
	for _, workflow := range svc.workflowsOrdered {
		svc.genTestClientImplWorkflowMethod(f, workflow)
		svc.genTestClientImplWorkflowAsyncMethod(f, workflow)
		svc.genTestClientImplWorkflowGetMethod(f, workflow)
		for _, signal := range svc.workflows[workflow].GetSignal() {
			if !signal.GetStart() {
				continue
			}
			svc.genTestClientImplWorkflowWithSignalMethod(f, workflow, signal.GetRef())
			svc.genTestClientImplWorkflowWithSignalAsyncMethod(f, workflow, signal.GetRef())
		}
	}

	// generate test client query methods
	for _, query := range svc.queriesOrdered {
		svc.genTestClientImplQueryMethod(f, query)
	}

	// generate test client signal methods
	for _, signal := range svc.signalsOrdered {
		svc.genTestClientImplSignalMethod(f, signal)
	}

	// generate test client update methods
	for _, update := range svc.updatesOrdered {
		svc.genTestClientImplUpdateMethod(f, update)
		svc.genTestClientImplUpdateAsyncMethod(f, update)
	}

	// generate workflow test runs
	for _, workflow := range svc.workflowsOrdered {
		opts := svc.workflows[workflow]
		svc.genTestClientWorkflowRunImpl(f, workflow)
		svc.genTestClientWorkflowRunImplGetMethod(f, workflow)
		svc.genTestClientWorkflowRunImplIDMethod(f, workflow)
		svc.genTestClientWorkflowRunImplRunIDMethod(f, workflow)

		// generate query methods
		for _, queryOpts := range opts.GetQuery() {
			svc.genTestClientWorkflowRunImplQueryMethod(f, workflow, queryOpts.GetRef())
		}

		// generate signal methods
		for _, signalOpts := range opts.GetSignal() {
			svc.genTestClientWorkflowRunImplSignalMethod(f, workflow, signalOpts.GetRef())
		}

		// generate update methods
		for _, updateOpts := range opts.GetUpdate() {
			svc.genTestClientWorkflowRunImplUpdateMethod(f, workflow, updateOpts.GetRef())
			svc.genTestClientWorkflowRunImplUpdateAsyncMethod(f, workflow, updateOpts.GetRef())
		}
	}
}
