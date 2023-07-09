package plugin

import (
	"fmt"
	"strings"

	g "github.com/dave/jennifer/jen"
)

// genActivitiesInterface generates an Activities interface
func (svc *Service) genActivitiesInterface(f *g.File) {
	typeName := svc.Names().ActivitiesInterface()
	f.Commentf("%s describes available worker activites", typeName)
	f.Type().Id(typeName).InterfaceFunc(func(methods *g.Group) {
		// define activity methods
		for _, activity := range svc.activitiesOrdered {
			method := svc.methods[activity]
			hasInput := !isEmpty(method.Input)
			hasOutput := !isEmpty(method.Output)
			methodName := svc.Names().ActivityGoName(activity)
			activityName := svc.Names().ActivityName(activity)

			desc := method.Comments.Leading.String()
			if desc != "" {
				desc = strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(desc, "//"), "\n//", ""))
			} else {
				desc = fmt.Sprintf("%s executes a %s signal", methodName, activityName)
			}

			methods.Comment(desc)
			methods.Id(methodName).
				ParamsFunc(func(args *g.Group) {
					args.Id("ctx").Qual("context", "Context")
					if hasInput {
						args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
					}
				}).
				ParamsFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Op("*").Id(method.Output.GoIdent.GoName)
					}
					returnVals.Error()
				})
		}
	})
}

// genActivitiesInterface generates a RegisterActivities public function
func (svc *Service) genRegisterActivities(f *g.File) {
	methodName := svc.Names().RegisterActivities()
	interfaceName := svc.Names().ActivitiesInterface()
	f.Commentf("%s registers %s activities with a worker", methodName, svc.GoName)
	f.Func().Id(methodName).
		Params(
			g.Id("r").Qual(workerPkg, "Registry"),
			g.Id("activities").Id(interfaceName),
		).
		BlockFunc(func(fn *g.Group) {
			for _, activity := range svc.activitiesOrdered {
				methodName = svc.Names().ActivityGoName(activity)
				fn.Id(methodName).Call(
					g.Id("r"), g.Id("activities").Dot(methodName),
				)
			}
		})
}

// genRegisterActivity generates a Register<Activity> public function
func (svc *Service) genRegisterActivity(f *g.File, activity string) {
	method := svc.methods[activity]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	methodName := svc.Names().RegisterActivitiy(activity)
	activityName := svc.Names().ActivityNameConstant(activity)

	f.Commentf("%s registers a %s activity", methodName, svc.Names().ActivityName(activity))
	f.Func().Id(methodName).
		Params(
			g.Id("r").Qual(workerPkg, "Registry"),
			g.Id("fn").Func().
				ParamsFunc(func(args *g.Group) {
					args.Qual("context", "Context")
					if hasInput {
						args.Op("*").Id(method.Input.GoIdent.GoName)
					}
				}).
				ParamsFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Op("*").Id(method.Output.GoIdent.GoName)
					}
					returnVals.Error()
				}),
		).
		Block(
			g.Id("r").Dot("RegisterActivityWithOptions").Call(
				g.Id("fn"), g.Qual(activityPkg, "RegisterOptions").Block(
					g.Id("Name").Op(":").Id(activityName).Op(","),
				),
			),
		)
}

// genActivityFuture generates a <Activity>Future struct
func (svc *Service) genActivityFuture(f *g.File, activity string) {
	future := svc.Names().ActivityFutureImpl(activity)

	f.Commentf("%s describes a %s activity execution", future, activity)
	f.Type().Id(future).Struct(
		g.Id("Future").Qual(workflowPkg, "Future"),
	)
}

// genActivityFutureGetMethod generates a <Workflow>Future's Get method
func (svc *Service) genActivityFutureGetMethod(f *g.File, activity string) {
	method := svc.methods[activity]
	hasOutput := !isEmpty(method.Output)
	future := svc.Names().ActivityFutureImpl(activity)

	f.Comment("Get blocks on activity execution, returning the response")
	f.Func().
		Params(g.Id("f").Op("*").Id(future)).
		Id("Get").
		Params(g.Id("ctx").Qual(workflowPkg, "Context")).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		BlockFunc(func(fn *g.Group) {
			if hasOutput {
				fn.Var().Id("resp").Id(method.Output.GoIdent.GoName)
				fn.If(
					g.Err().Op(":=").Id("f").Dot("Future").Dot("Get").Call(
						g.Id("ctx"), g.Op("&").Id("resp"),
					),
					g.Err().Op("!=").Nil(),
				).Block(
					g.Return(g.Nil(), g.Err()),
				)
				fn.Return(g.Op("&").Id("resp"), g.Nil())
			} else {
				fn.Return(g.Id("f").Dot("Future").Dot("Get").Call(
					g.Id("ctx"), g.Nil(),
				))
			}
		})
}

// genActivityFutureSelectMethod generates a <Workflow>Future's Select method
func (svc *Service) genActivityFutureSelectMethod(f *g.File, activity string) {
	future := svc.Names().ActivityFutureImpl(activity)

	f.Comment("Select adds the activity completion to the selector, callback can be nil")
	f.Func().
		Params(g.Id("f").Op("*").Id(future)).
		Id("Select").
		Params(
			g.Id("sel").Qual(workflowPkg, "Selector"),
			g.Id("fn").Func().Params(g.Op("*").Id(future)),
		).
		Params(
			g.Qual(workflowPkg, "Selector"),
		).
		Block(
			g.Return(
				g.Id("sel").Dot("AddFuture").Call(
					g.Id("f").Dot("Future"),
					g.Func().
						Params(g.Qual(workflowPkg, "Future")).
						Block(
							g.If(g.Id("fn").Op("!=").Nil()).Block(
								g.Id("fn").Call(g.Id("f")),
							),
						),
				),
			),
		)
}
