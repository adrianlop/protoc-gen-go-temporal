package plugin

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	g "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

// // genWorkerWorkflowFactoryExecuteMethod generates a <Workflow>Worker's <Workflow> method
// func (svc *Service) genWorkerWorkflowFactoryExecuteMethod(f *g.File, workflow string) {
// 	method := svc.methods[workflow]
// 	opts := svc.workflows[workflow]
// 	hasInput := !isEmpty(method.Input)
// 	hasOutput := !isEmpty(method.Output)
// 	privateName := pgs.Name(method.GoName).LowerCamelCase().String()
// 	workerName := privateName

// 	// generate <Workflow> method for worker struct
// 	f.Commentf("%s constructs a new %s value and executes it", method.GoName, method.GoName)
// 	f.Func().
// 		Params(g.Id("w").Op("*").Id(workerName)).
// 		Id(method.GoName).
// 		ParamsFunc(func(args *g.Group) {
// 			args.Id("ctx").Qual(workflowPkg, "Context")
// 			if hasInput {
// 				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
// 			}
// 		}).
// 		ParamsFunc(func(returnVals *g.Group) {
// 			if hasOutput {
// 				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
// 			}
// 			returnVals.Error()
// 		}).
// 		BlockFunc(func(fn *g.Group) {
// 			// build input struct
// 			fn.Id("input").Op(":=").Op("&").Id(fmt.Sprintf("%sInput", method.GoName)).BlockFunc(func(fields *g.Group) {
// 				if hasInput {
// 					fields.Id("Req").Op(":").Id("req").Op(",")
// 				}
// 				for _, s := range opts.GetSignal() {
// 					signal := s.GetRef()
// 					fields.Id(signal).Op(":").Op("&").Id(fmt.Sprintf("%sSignal", signal)).Block(
// 						g.Id("Channel").Op(":").Qual(workflowPkg, "GetSignalChannel").Call(
// 							g.Id("ctx"), g.Id(fmt.Sprintf("%sSignalName", signal)),
// 						).Op(","),
// 					).Op(",")
// 				}
// 			})

// 			// call constructor to get workflow implementation
// 			fn.List(g.Id("wf"), g.Err()).Op(":=").Id("w").Dot("ctor").Call(
// 				g.Id("ctx"), g.Id("input"),
// 			)
// 			fn.If(g.Err().Op("!=").Nil()).Block(
// 				g.ReturnFunc(func(returnVals *g.Group) {
// 					if hasOutput {
// 						returnVals.Nil()
// 					}
// 					returnVals.Err()
// 				}),
// 			)

// 			// register query handlers
// 			for _, q := range opts.GetQuery() {
// 				query := q.GetRef()
// 				fn.If(
// 					g.Err().Op(":=").Qual(workflowPkg, "SetQueryHandler").Call(
// 						g.Id("ctx"), g.Id(fmt.Sprintf("%sQueryName", query)), g.Id("wf").Dot(query),
// 					),
// 					g.Err().Op("!=").Nil(),
// 				).Block(
// 					g.ReturnFunc(func(returnVals *g.Group) {
// 						if hasOutput {
// 							returnVals.Nil()
// 						}
// 						returnVals.Err()
// 					}),
// 				)
// 			}

// 			// register update handlers
// 			for _, u := range opts.GetUpdate() {
// 				update := u.GetRef()
// 				updateOpts := svc.updates[update]
// 				updateHandlerOptionsName := fmt.Sprintf("%sOpts", pgs.Name(update).LowerCamelCase().String())

// 				// build UpdateHandlerOptions
// 				var updateHandlerOptions []g.Code
// 				if updateOpts.GetValidate() {
// 					updateHandlerOptions = append(updateHandlerOptions, g.Id("Validator").Op(":").Id("wf").Dot(fmt.Sprintf("Validate%s", update)))
// 				}
// 				fn.Id(updateHandlerOptionsName).Op(":=").Qual(workflowPkg, "UpdateHandlerOptions").Values(updateHandlerOptions...)

// 				fn.If(
// 					g.Err().Op(":=").Qual(workflowPkg, "SetUpdateHandlerWithOptions").Call(
// 						g.Id("ctx"), g.Id(fmt.Sprintf("%sUpdateName", update)), g.Id("wf").Dot(update), g.Id(updateHandlerOptionsName),
// 					),
// 					g.Err().Op("!=").Nil(),
// 				).Block(
// 					g.ReturnFunc(func(returnVals *g.Group) {
// 						if hasOutput {
// 							returnVals.Nil()
// 						}
// 						returnVals.Err()
// 					}),
// 				)
// 			}

// 			// execute workflow
// 			fn.Return(
// 				g.Id("wf").Dot("Execute").Call(g.Id("ctx")),
// 			)
// 		})
// }

// genWorkerRegisterWorkflow generates a Register<Workflow> public function
func (svc *Service) genWorkerRegisterWorkflow(f *g.File, workflow string) {
	// generate Register<Workflow> function
	f.Commentf("%s registers a %s workflow with the given worker", svc.Names().RegisterWorkflow(workflow), svc.Names().WorkflowName(workflow))
	f.Func().
		Id(svc.Names().RegisterWorkflow(workflow)).
		Params(
			g.Id("r").Qual(workerPkg, "Registry"),
			g.Id("wf").
				Func().
				Params(
					g.Qual(workflowPkg, "Context"),
					g.Op("*").Id(svc.Names().WorkflowInput(workflow)),
				).
				Params(
					g.Id(svc.Names().WorkflowInterface(workflow)),
					g.Error(),
				),
			// g.Id("wf").Id("Workflows"),
		).
		Block(
			g.Id("r").Dot("RegisterWorkflowWithOptions").Call(
				g.Id(svc.Names().Builder(workflow)).Call(g.Id("wf")),
				g.Qual(workflowPkg, "RegisterOptions").Values(
					g.Id("Name").Op(":").Id(svc.Names().WorkflowNameConstant(workflow)),
				),
			),
		)
}

// genWorkerRegisterWorkflows generates a public RegisterWorkflows method for a given service
func (svc *Service) genWorkerRegisterWorkflows(f *g.File) {
	workflowsInterface := svc.Names().WorkflowsInterface()
	// generate workflow registration function for service
	f.Commentf("%s registers %s workflows with the given worker", svc.Names().RegisterWorkflows(), svc.GoName)
	f.Func().
		Id(svc.Names().RegisterWorkflows()).
		Params(
			g.Id("r").Qual(workerPkg, "Registry"),
			g.Id("workflows").Id(workflowsInterface),
		).
		BlockFunc(func(fn *g.Group) {
			for _, workflow := range svc.workflowsOrdered {
				workflowMethod := svc.Names().WorkflowGoName(workflow)
				fn.Id(svc.Names().RegisterWorkflow(workflow)).Call(
					g.Id("r"), g.Id("workflows").Dot(workflowMethod),
				)
			}
		})
}

// genWorkerSignal generates a worker signal struct
func (svc *Service) genWorkerSignal(f *g.File, signal string) {
	signalType := svc.Names().SignalImpl(signal)
	f.Commentf("%s describes a %s signal", signalType, svc.Names().SignalName(signal))
	f.Type().Id(signalType).Struct(
		g.Id("Channel").Qual(workflowPkg, "ReceiveChannel"),
	)
}

// genWorkerSignalReceive generates a worker signal Receive method
func (svc *Service) genWorkerSignalReceive(f *g.File, signal string) {
	method := svc.methods[signal]
	hasInput := !isEmpty(method.Input)
	signalType := svc.Names().SignalImpl(signal)
	signalName := svc.Names().SignalName(signal)
	f.Commentf("Receive blocks until a %s signal is received", signalName)
	f.Func().
		Params(g.Id("s").Op("*").Id(signalType)).
		Id("Receive").
		Params(g.Id("ctx").Qual(workflowPkg, "Context")).
		ParamsFunc(func(returnVals *g.Group) {
			if hasInput {
				returnVals.Op("*").Id(method.Input.GoIdent.GoName)
			}
			returnVals.Bool()
		}).
		BlockFunc(func(b *g.Group) {
			if hasInput {
				b.Var().Id("resp").Id(method.Input.GoIdent.GoName)
			}
			b.Id("more").Op(":=").Id("s").Dot("Channel").Dot("Receive").CallFunc(func(args *g.Group) {
				args.Id("ctx")
				if hasInput {
					args.Op("&").Id("resp")
				} else {
					args.Nil()
				}
			})
			b.ReturnFunc(func(returnVals *g.Group) {
				if hasInput {
					returnVals.Op("&").Id("resp")
				}
				returnVals.Id("more")
			})
		})
}

// genWorkerSignalReceiveAsync generates a worker signal ReceiveAsync method
func (svc *Service) genWorkerSignalReceiveAsync(f *g.File, signal string) {
	method := svc.methods[signal]
	hasInput := !isEmpty(method.Input)
	signalType := svc.Names().SignalImpl(signal)
	signalName := svc.Names().SignalName(signal)
	f.Commentf("ReceiveAsync checks for a %s signal without blocking", signalName)
	f.Func().
		Params(g.Id("s").Op("*").Id(signalType)).
		Id("ReceiveAsync").
		Params().
		ParamsFunc(func(returnVals *g.Group) {
			if hasInput {
				returnVals.Op("*").Id(method.Input.GoIdent.GoName)
			} else {
				returnVals.Bool()
			}
		}).
		BlockFunc(func(b *g.Group) {
			if hasInput {
				b.Var().Id("resp").Id(method.Input.GoIdent.GoName)
				b.If(
					g.Id("ok").Op(":=").Id("s").Dot("Channel").Dot("ReceiveAsync").Call(
						g.Op("&").Id("resp"),
					),
					g.Op("!").Id("ok"),
				).Block(
					g.Return(g.Nil()),
				)
				b.Return(g.Op("&").Id("resp"))
			} else {
				b.Return(g.Id("s").Dot("Channel").Dot("ReceiveAsync").Call(g.Nil()))
			}
		})
}

// genWorkerSignalSelect generates a worker signal Select method
func (svc *Service) genWorkerSignalSelect(f *g.File, signal string) {
	method := svc.methods[signal]
	hasInput := !isEmpty(method.Input)
	signalType := svc.Names().SignalImpl(signal)
	signalName := svc.Names().SignalName(signal)
	f.Commentf("Select checks for a %s signal without blocking", signalName)
	f.Func().
		Params(g.Id("s").Op("*").Id(signalType)).
		Id("Select").
		Params(
			g.Id("sel").Qual(workflowPkg, "Selector"),
			g.Id("fn").Func().ParamsFunc(func(args *g.Group) {
				if hasInput {
					args.Op("*").Id(method.Input.GoIdent.GoName)
				}
			}),
		).
		Params(
			g.Qual(workflowPkg, "Selector"),
		).
		Block(
			g.Return(
				g.Id("sel").Dot("AddReceive").Call(
					g.Id("s").Dot("Channel"),
					g.Func().
						Params(
							g.Qual(workflowPkg, "ReceiveChannel"),
							g.Bool(),
						).
						BlockFunc(func(fn *g.Group) {
							if hasInput {
								fn.Id("req").Op(":=").Id("s").Dot("ReceiveAsync").Call()
							} else {
								fn.Id("s").Dot("ReceiveAsync").Call()
							}
							fn.If(g.Id("fn").Op("!=").Nil()).Block(
								g.Id("fn").CallFunc(func(args *g.Group) {
									if hasInput {
										args.Id("req")
									}
								}),
							)
						}),
				),
			),
		)
}

// genWorkerWorkflowBuilder generates a build<Workflow> function that converts
// a constructor function or method into a valid workflow function
func (svc *Service) genWorkerWorkflowBuilder(f *g.File, workflow string) {
	method := svc.methods[workflow]
	opts := svc.workflows[workflow]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	functionName := svc.Names().Builder(workflow)
	structName := svc.Names().WorkflowImpl(workflow)

	// generate Build<Workflow> function
	f.Commentf("%s converts a %s struct into a valid workflow function", functionName, structName)
	f.Func().
		Id(functionName).
		Params(
			g.Id("wf").
				Func().
				Params(
					g.Qual(workflowPkg, "Context"),
					g.Op("*").Id(svc.Names().WorkflowInput(workflow)),
				).
				Params(
					g.Id(svc.Names().WorkflowInterface(workflow)),
					g.Error(),
				),
		).
		Params(
			g.Func().
				ParamsFunc(func(args *g.Group) {
					args.Qual(workflowPkg, "Context")
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
			g.Return(
				//g.Parens(g.Op("&").Id(svc.Names().WorkflowImpl(workflow)).Values(g.Id("wf"))).Dot(method.GoName),
				g.Func().
					ParamsFunc(func(args *g.Group) {
						args.Id("ctx").Qual(workflowPkg, "Context")
						if hasInput {
							args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
						}
					}).
					ParamsFunc(func(returnVals *g.Group) {
						if hasOutput {
							returnVals.Op("*").Id(method.Output.GoIdent.GoName)
						}
						returnVals.Error()
					}).
					BlockFunc(func(fn *g.Group) {
						// build input struct
						fn.Id("input").Op(":=").Op("&").Id(svc.Names().WorkflowInput(workflow)).CustomFunc(multiLineValues, func(fields *g.Group) {
							if hasInput {
								fields.Id("Req").Op(":").Id("req")
							}
							for _, s := range opts.GetSignal() {
								signal := s.GetRef()
								fields.Id(signal).Op(":").Op("&").Id(svc.Names().SignalImpl(signal)).Custom(multiLineValues,
									g.Id("Channel").Op(":").Qual(workflowPkg, "GetSignalChannel").Call(
										g.Id("ctx"), g.Id(svc.Names().SignalNameConstant(signal)),
									),
								)
							}
						})

						// call constructor to get workflow implementation
						fn.List(g.Id("wf"), g.Err()).Op(":=").Id("wf").Call(
							g.Id("ctx"), g.Id("input"),
						)
						fn.If(g.Err().Op("!=").Nil()).Block(
							g.ReturnFunc(func(returnVals *g.Group) {
								if hasOutput {
									returnVals.Nil()
								}
								returnVals.Err()
							}),
						)

						// register query handlers
						for _, q := range opts.GetQuery() {
							query := q.GetRef()
							fn.If(
								g.Err().Op(":=").Qual(workflowPkg, "SetQueryHandler").Call(
									g.Id("ctx"), g.Id(svc.Names().QueryNameConstant(query)), g.Id("wf").Dot(query),
								),
								g.Err().Op("!=").Nil(),
							).Block(
								g.ReturnFunc(func(returnVals *g.Group) {
									if hasOutput {
										returnVals.Nil()
									}
									returnVals.Err()
								}),
							)
						}

						// register update handlers
						for _, u := range opts.GetUpdate() {
							update := u.GetRef()
							updateOpts := svc.updates[update]

							fn.Commentf("Register %q update handler", svc.Names().UpdateName(update))
							fn.BlockFunc(func(bl *g.Group) {
								// build UpdateHandlerOptions
								var updateHandlerOptions []g.Code
								if updateOpts.GetValidate() {
									updateHandlerOptions = append(updateHandlerOptions, g.Id("Validator").Op(":").Id("wf").Dot(fmt.Sprintf("Validate%s", update)))
								}
								fn.Id("opts").Op(":=").Qual(workflowPkg, "UpdateHandlerOptions").Values(updateHandlerOptions...)

								fn.If(
									g.Err().Op(":=").Qual(workflowPkg, "SetUpdateHandlerWithOptions").Call(
										g.Id("ctx"), g.Id(svc.Names().UpdateNameConstant(update)), g.Id("wf").Dot(update), g.Id("opts"),
									),
									g.Err().Op("!=").Nil(),
								).Block(
									g.ReturnFunc(func(returnVals *g.Group) {
										if hasOutput {
											returnVals.Nil()
										}
										returnVals.Err()
									}),
								)
							})

						}

						// execute workflow
						fn.Return(
							g.Id("wf").Dot("Execute").CallFunc(func(args *g.Group) {
								args.Id("ctx")
								if hasInput {
									args.Id("req")
								}
							}),
						)
					}),
			),
		)
}

// genWorkerWorkflowChildRunImpl generates a <Workflow>ChildRun struct
func (svc *Service) genWorkerWorkflowChildRunImpl(f *g.File, workflow string) {
	// generate child workflow run struct
	typeName := svc.Names().WorkflowChildRunImpl(workflow)
	workflowName := svc.Names().WorkflowName(workflow)
	f.Commentf("%sChildRun describes a child %s workflow run", typeName, workflowName)
	f.Type().Id(typeName).StructFunc(func(fields *g.Group) {
		fields.Add(g.Id("Future").Qual(workflowPkg, "ChildWorkflowFuture"))
	})
}

// genWorkerWorkflowChildRunImplGetMethod generates a <Workflow>ChildRun Get method
func (svc *Service) genWorkerWorkflowChildRunImplGetMethod(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasOutput := !isEmpty(method.Output)
	typeName := svc.Names().WorkflowChildRunImpl(workflow)
	f.Comment("Get blocks until the workflow is completed, returning the response value")
	f.Func().
		Params(
			g.Id("r").Op("*").Id(typeName),
		).
		Id("Get").
		Params(
			g.Id("ctx").Qual(workflowPkg, "Context"),
		).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		BlockFunc(func(fn *g.Group) {
			if hasOutput {
				fn.Var().Id("resp").Id(method.Output.GoIdent.GoName)
			}
			fn.If(
				g.Err().Op(":=").Id("r").Dot("Future").Dot("Get").CallFunc(func(args *g.Group) {
					args.Id("ctx")
					if hasOutput {
						args.Op("&").Id("resp")
					} else {
						args.Nil()
					}
				}),
				g.Err().Op("!=").Nil(),
			).BlockFunc(func(b *g.Group) {
				if hasOutput {
					b.Return(g.Nil(), g.Err())
				} else {
					b.Return(g.Err())
				}
			})
			if hasOutput {
				fn.Return(g.Op("&").Id("resp"), g.Nil())
			} else {
				fn.Return(g.Nil())
			}
		})
}

// genWorkerWorkflowChildRunImplSelectMethod generates a <Workflow>ChildRun Select method
func (svc *Service) genWorkerWorkflowChildRunImplSelectMethod(f *g.File, workflow string) {
	interfaceName := svc.Names().WorkflowChildRunInterface(workflow)
	typeName := svc.Names().WorkflowChildRunImpl(workflow)
	f.Comment("Select adds this completion to the selector. Callback can be nil.")
	f.Func().
		Params(
			g.Id("r").Op("*").Id(typeName),
		).
		Id("Select").
		Params(
			g.Id("sel").Qual(workflowPkg, "Selector"),
			g.Id("fn").Func().Params(g.Id(interfaceName)),
		).
		Params(
			g.Qual(workflowPkg, "Selector"),
		).
		Block(
			g.Return(
				g.Id("sel").Dot("AddFuture").Call(
					g.Id("r").Dot("Future"),
					g.Func().Params(g.Qual(workflowPkg, "Future")).Block(
						g.If(g.Id("fn").Op("!=").Nil()).Block(
							g.Id("fn").Call(g.Id("r")),
						),
					),
				),
			),
		)
}

// genWorkerWorkflowChildRunImplSelectStartMethod generates a <Workflow>ChildRun SelectStart method
func (svc *Service) genWorkerWorkflowChildRunImplSelectStartMethod(f *g.File, workflow string) {
	interfaceName := svc.Names().WorkflowChildRunInterface(workflow)
	typeName := svc.Names().WorkflowChildRunImpl(workflow)
	f.Comment("SelectStart adds waiting for start to the selector. Callback can be nil.")
	f.Func().
		Params(
			g.Id("r").Op("*").Id(typeName),
		).
		Id("SelectStart").
		Params(
			g.Id("sel").Qual(workflowPkg, "Selector"),
			g.Id("fn").Func().Params(g.Id(interfaceName)),
		).
		Params(
			g.Qual(workflowPkg, "Selector"),
		).
		Block(
			g.Return(
				g.Id("sel").Dot("AddFuture").Call(
					g.Id("r").Dot("Future").Dot("GetChildWorkflowExecution").Call(),
					g.Func().Params(g.Qual(workflowPkg, "Future")).Block(
						g.If(g.Id("fn").Op("!=").Nil()).Block(
							g.Id("fn").Call(g.Id("r")),
						),
					),
				),
			),
		)
}

// genWorkerWorkflowChildRunImplSignalMethods generates <Workflow>ChildRun signal methods
func (svc *Service) genWorkerWorkflowChildRunImplSignalMethods(f *g.File, workflow string) {
	opts := svc.workflows[workflow]
	typeName := svc.Names().WorkflowChildRunImpl(workflow)
	for _, signalOpts := range opts.GetSignal() {
		signal := signalOpts.GetRef()
		handler := svc.methods[signal]
		hasInput := !isEmpty(handler.Input)
		methodName := svc.Names().SignalGoName(signal)
		signalName := svc.Names().SignalNameConstant(workflow)

		f.Commentf("%s sends a %q signal request to the child workflow", methodName, signalName)
		f.Func().
			Params(g.Id("r").Op("*").Id(typeName)).
			Id(methodName).
			ParamsFunc(func(params *g.Group) {
				params.Id("ctx").Qual(workflowPkg, "Context")
				if hasInput {
					params.Id("input").Op("*").Id(handler.Input.GoIdent.GoName)
				}
			}).
			Params(g.Qual(workflowPkg, "Future")).
			Block(
				g.Return(g.Id("r").Dot("Future").Dot("SignalChildWorkflow").CallFunc(func(args *g.Group) {
					args.Id("ctx")
					args.Id(signalName)
					if hasInput {
						args.Id("input")
					} else {
						args.Nil()
					}
				})),
			)
	}
}

// genWorkerWorkflowChildRunIpmlWaitStartMethod generates a <Workflow>ChildRun WaitStart method
func (svc *Service) genWorkerWorkflowChildRunIpmlWaitStartMethod(f *g.File, workflow string) {
	typeName := svc.Names().WorkflowChildRunImpl(workflow)
	f.Comment("WaitStart waits for the child workflow to start")
	f.Func().
		Params(
			g.Id("r").Op("*").Id(typeName),
		).
		Id("WaitStart").
		Params(
			g.Id("ctx").Qual(workflowPkg, "Context"),
		).
		Params(
			g.Op("*").Qual(workflowPkg, "Execution"),
			g.Error(),
		).
		Block(
			g.Var().Id("exec").Qual(workflowPkg, "Execution"),
			g.If(
				g.Err().Op(":=").Id("r").Dot("Future").Dot("GetChildWorkflowExecution").Call().Dot("Get").Call(
					g.Id("ctx"),
					g.Op("&").Id("exec"),
				),
				g.Err().Op("!=").Nil(),
			).Block(
				g.Return(g.Nil(), g.Err()),
			),
			g.Return(g.Op("&").Id("exec"), g.Nil()),
		)
}

// // genWorkerWorkflowFactory generates a <Workflow> struct that is used by the builder
// func (svc *Service) genWorkerWorkflowFactory(f *g.File, workflow string) {
// 	method := svc.methods[workflow]
// 	privateName := pgs.Name(method.GoName).LowerCamelCase().String()
// 	workerName := privateName
// 	f.Commentf("%s provides an %s method for calling the user's implementation", workerName, method.GoName)
// 	f.Type().
// 		Id(workerName).
// 		Struct(
// 			g.Id("ctor").
// 				Func().
// 				Params(
// 					g.Qual(workflowPkg, "Context"),
// 					g.Op("*").Id(fmt.Sprintf("%sInput", method.GoName)),
// 				).
// 				Params(
// 					g.Id(fmt.Sprintf("%sWorkflow", method.GoName)),
// 					g.Error(),
// 				),
// 		)
// }

// genWorkerWorkflowInput generates a <Workflow>Input struct
func (svc *Service) genWorkerWorkflowInput(f *g.File, workflow string) {
	opts := svc.workflows[workflow]
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	typeName := svc.Names().WorkflowInput(workflow)
	f.Commentf("%sInput describes the input to a %s workflow constructor", workflow, workflow)
	f.Type().Id(typeName).StructFunc(func(fields *g.Group) {
		fields.Op("*").Id(svc.Names().WorkflowResources())
		if hasInput {
			fields.Id("Req").Op("*").Id(method.Input.GoIdent.GoName)
		}

		// add workflow signals
		for _, signalOpts := range opts.GetSignal() {
			signal := signalOpts.GetRef()
			fields.Id(svc.Names().SignalGoName(signal)).Op("*").Id(svc.Names().SignalImpl(signal))
		}
	})
}

// genWorkerWorkflowInterface generates a <Workflow> interface
func (svc *Service) genWorkerWorkflowInterface(f *g.File, workflow string) {
	opts := svc.workflows[workflow]
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	typeName := svc.Names().WorkflowInterface(workflow)
	workflowName := svc.Names().WorkflowName(workflow)
	// generate workflow interface
	if method.Comments.Leading.String() != "" {
		f.Comment(strings.TrimSuffix(method.Comments.Leading.String(), "\n"))
	} else {
		f.Commentf("%s describes a %q workflow implementation", typeName, workflowName)
	}
	f.Type().Id(typeName).InterfaceFunc(func(methods *g.Group) {
		methods.Commentf("Execute a %s workflow", workflowName)
		methods.Id("Execute").
			ParamsFunc(func(args *g.Group) {
				args.Id("ctx").Qual(workflowPkg, "Context")
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

		// add workflow query methods
		for _, queryOpts := range opts.GetQuery() {
			query := queryOpts.GetRef()
			handler := svc.methods[query]
			hasInput := !isEmpty(handler.Input)
			methodName := svc.Names().QueryGoName(query)
			queryName := svc.Names().QueryName(query)

			methods.Commentf("%s implements a(n) %q query handler", methodName, queryName)
			methods.Id(methodName).
				ParamsFunc(func(args *g.Group) {
					if hasInput {
						args.Op("*").Id(handler.Input.GoIdent.GoName)
					}
				}).
				Params(
					g.Op("*").Id(handler.Output.GoIdent.GoName),
					g.Error(),
				)
		}

		// add workflow update methods
		for _, updateOpts := range opts.GetUpdate() {
			update := updateOpts.GetRef()
			handler := svc.methods[update]
			handlerOpts := svc.updates[update]
			hasInput := !isEmpty(handler.Input)
			hasOutput := !isEmpty(handler.Output)
			methodName := svc.Names().QueryGoName(update)
			updateName := svc.Names().QueryName(update)

			// add Validate<Update> method if enabled
			if handlerOpts.GetValidate() {
				validatorName := fmt.Sprintf("Validate%s", methodName)
				methods.Commentf("%s validates a(n) %s update", validatorName, updateName)
				methods.Id(validatorName).
					ParamsFunc(func(args *g.Group) {
						args.Qual(workflowPkg, "Context")
						if hasInput {
							args.Op("*").Id(handler.Input.GoIdent.GoName)
						}
					}).
					Params(g.Error())
			}

			// add <Update> method
			if handler.Comments.Leading.String() != "" {
				methods.Comment(strings.TrimSuffix(handler.Comments.Leading.String(), "\n"))
			} else {
				methods.Commentf("%s implements a(n) %q update handler", methodName, updateName)
			}
			methods.Id(methodName).
				ParamsFunc(func(args *g.Group) {
					args.Qual(workflowPkg, "Context")
					if hasInput {
						args.Op("*").Id(handler.Input.GoIdent.GoName)
					}
				}).
				ParamsFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Op("*").Id(handler.Output.GoIdent.GoName)
					}
					returnVals.Error()
				})
		}
	})
}

// genWorkerWorkflowResources generates an embeddable <Workflow>Resources struct that provides
// activity, external signal, child workflow methods
func (svc *Service) genWorkerWorkflowResources(f *g.File) {
	f.Commentf("%s provides convenience methods for use within worfklows", svc.Names().WorkflowResources())
	f.Type().Id(svc.Names().WorkflowResources()).Struct()
}

// genWorkerWorkflowResourcesActivityMethod generates a <Activity>[Local] method
func (svc *Service) genWorkerWorkflowResourcesActivityMethod(f *g.File, activity string, local, async bool) {
	method := svc.methods[activity]
	opts := svc.activities[activity]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	activityName := svc.Names().ActivityName(activity)
	futureType := svc.Names().ActivityFutureImpl(activity)

	methodName := svc.Names().ActivityGoName(activity)
	var annotations []string
	if local {
		methodName = fmt.Sprintf("%sLocal", methodName)
		annotations = append(annotations, "locally")
	}
	if async {
		methodName = fmt.Sprintf("%sAsync", methodName)
		annotations = append(annotations, "asynchronously")
	}
	sort.Slice(annotations, func(i, j int) bool {
		return annotations[i] < annotations[j]
	})

	desc := method.Comments.Leading.String()
	if desc != "" {
		desc = strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(desc, "//"), "\n//", ""))
	} else {
		desc = fmt.Sprintf("%s executes a(n) %s activity", methodName, activityName)
	}
	if len(annotations) > 0 {
		desc = fmt.Sprintf("%s (%s)", desc, strings.Join(annotations, ", "))
	}

	f.Comment(desc)
	f.Func().
		Params(g.Id("r").Op("*").Id(svc.Names().WorkflowResources())).
		Id(methodName).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual(workflowPkg, "Context")
			if local {
				args.Id("fn").
					Func().
					ParamsFunc(func(fnargs *g.Group) {
						fnargs.Qual("context", "Context")
						if hasInput {
							fnargs.Op("*").Id(method.Input.GoIdent.GoName)
						}
					}).
					ParamsFunc(func(fnreturn *g.Group) {
						if hasOutput {
							fnreturn.Op("*").Id(method.Output.GoIdent.GoName)
						}
						fnreturn.Error()
					})
			}
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			if local {
				args.Id("options").Op("...").Op("*").Qual(workflowPkg, "LocalActivityOptions")
			} else {
				args.Id("options").Op("...").Op("*").Qual(workflowPkg, "ActivityOptions")
			}
		}).
		ParamsFunc(func(returnVals *g.Group) {
			if async {
				returnVals.Op("*").Id(futureType)
			} else {
				if hasOutput {
					returnVals.Op("*").Id(method.Output.GoIdent.GoName)
				}
				returnVals.Error()
			}
		}).
		BlockFunc(func(fn *g.Group) {
			// initialize activity options if nil
			if local {
				fn.Var().Id("opts").Op("*").Qual(workflowPkg, "LocalActivityOptions")
			} else {
				fn.Var().Id("opts").Op("*").Qual(workflowPkg, "ActivityOptions")
			}
			fn.If(g.Len(g.Id("options")).Op(">").Lit(0).Op("&&").Id("options").Index(g.Lit(0)).Op("!=").Nil()).
				Block(
					g.Id("opts").Op("=").Id("options").Index(g.Lit(0)),
				).
				Else().
				BlockFunc(func(bl *g.Group) {
					optionsFn := "GetActivityOptions"
					if local {
						optionsFn = "GetLocalActivityOptions"
					}
					bl.Id("activityOpts").Op(":=").Qual(workflowPkg, optionsFn).Call(
						g.Id("ctx"),
					)
					bl.Id("opts").Op("=").Op("&").Id("activityOpts")
				})

			// set default retry policy
			if policy := opts.GetRetryPolicy(); policy != nil {
				fn.If(g.Id("opts").Dot("RetryPolicy").Op("==").Nil()).Block(
					g.Id("opts").Dot("RetryPolicy").Op("=").Op("&").Qual(temporalPkg, "RetryPolicy").ValuesFunc(func(fields *g.Group) {
						if d := policy.GetInitialInterval(); d.IsValid() {
							fields.Id("InitialInterval").Op(":").Id(strconv.FormatInt(d.AsDuration().Nanoseconds(), 10))
						}
						if d := policy.GetMaxInterval(); d.IsValid() {
							fields.Id("MaximumInterval").Op(":").Id(strconv.FormatInt(d.AsDuration().Nanoseconds(), 10))
						}
						if n := policy.GetBackoffCoefficient(); n != 0 {
							fields.Id("BackoffCoefficient").Op(":").Lit(n)
						}
						if n := policy.GetMaxAttempts(); n != 0 {
							fields.Id("MaximumAttempts").Op(":").Lit(n)
						}
						if errs := policy.GetNonRetryableErrorTypes(); len(errs) > 0 {
							fields.Id("NonRetryableErrorTypes").Op(":").Lit(errs)
						}
					}),
				)
			}

			// set default heartbeat timeout
			if timeout := opts.GetHeartbeatTimeout(); !local && timeout.IsValid() {
				fn.If(g.Id("opts").Dot("HeartbeatTimeout").Op("==").Lit(0)).Block(
					g.Id("opts").Dot("HeartbeatTimeout").Op("=").Id(strconv.FormatInt(timeout.AsDuration().Nanoseconds(), 10)).Comment(timeout.AsDuration().String()),
				)
			}

			// set default schedule to close timeout
			if timeout := opts.GetScheduleToCloseTimeout(); timeout.IsValid() {
				fn.If(g.Id("opts").Dot("ScheduleToCloseTimeout").Op("==").Lit(0)).Block(
					g.Id("opts").Dot("ScheduleToCloseTimeout").Op("=").Id(strconv.FormatInt(timeout.AsDuration().Nanoseconds(), 10)).Comment(timeout.AsDuration().String()),
				)
			}

			// set default schedule to start timeout
			if timeout := opts.GetScheduleToStartTimeout(); !local && timeout.IsValid() {
				fn.If(g.Id("opts").Dot("ScheduleToStartTimeout").Op("==").Lit(0)).Block(
					g.Id("opts").Dot("ScheduleToStartTimeout").Op("=").Id(strconv.FormatInt(timeout.AsDuration().Nanoseconds(), 10)).Comment(timeout.AsDuration().String()),
				)
			}

			// set default start to close timeout
			if timeout := opts.GetStartToCloseTimeout(); timeout.IsValid() {
				fn.If(g.Id("opts").Dot("StartToCloseTimeout").Op("==").Lit(0)).Block(
					g.Id("opts").Dot("StartToCloseTimeout").Op("=").Id(strconv.FormatInt(timeout.AsDuration().Nanoseconds(), 10)).Comment(timeout.AsDuration().String()),
				)
			}

			// inject ctx with activity options
			if local {
				fn.Id("ctx").Op("=").Qual(workflowPkg, "WithLocalActivityOptions").Call(
					g.Id("ctx"), g.Op("*").Id("opts"),
				)

			} else {
				fn.Id("ctx").Op("=").Qual(workflowPkg, "WithActivityOptions").Call(
					g.Id("ctx"), g.Op("*").Id("opts"),
				)
			}

			// initialize activity reference
			fn.Var().Id("activity").Any()
			if local {
				fn.If(g.Id("fn").Op("==").Nil()).
					Block(
						g.Id("activity").Op("=").Id(fmt.Sprintf("%sActivityName", activity)),
					).
					Else().
					Block(
						g.Id("activity").Op("=").Id("fn"),
					)
			} else {
				fn.Id("activity").Op("=").Id(fmt.Sprintf("%sActivityName", method.GoName))
			}

			// initialize activity future
			fn.Id("future").Op(":=").Op("&").Id(futureType).ValuesFunc(func(values *g.Group) {
				methodName := "ExecuteActivity"
				if local {
					methodName = "ExecuteLocalActivity"
				}

				values.Id("Future").Op(":").Qual(workflowPkg, methodName).CallFunc(func(args *g.Group) {
					args.Id("ctx")
					args.Id("activity")
					if hasInput {
						args.Id("req")
					}
				})
			})

			fn.ReturnFunc(func(returnVals *g.Group) {
				if async {
					returnVals.Add(g.Id("future"))
				} else {
					returnVals.Add(g.Id("future").Dot("Get").Call(g.Id("ctx")))
				}
			})
		})
}

// genWorkerWorkflowResourcesChildWorkflowAsyncMethod generates a public <Workflow>Child function
func (svc *Service) genWorkerWorkflowResourcesChildWorkflowAsyncMethod(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	name := strcase.ToCamel(workflow + "ChildAsync")

	f.Commentf("%s executes a child %s workflow", name, workflow)
	f.Func().
		Id(name).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual(workflowPkg, "Context")
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			args.Id("options").Op("...").Op("*").Qual(workflowPkg, "ChildWorkflowOptions")
		}).
		Params(
			g.Id(svc.Names().WorkflowChildRunInterface(workflow)),
			g.Error(),
		).
		BlockFunc(func(fn *g.Group) {
			// initialize child workflow options with default values
			svc.genClientStartWorkflowOptions(fn, workflow, true)

			fn.Id("ctx").Op("=").Qual(workflowPkg, "WithChildOptions").Call(g.Id("ctx"), g.Op("*").Id("opts"))
			fn.Return(
				g.Op("&").Id(svc.Names().WorkflowChildRunImpl(workflow)).Values(
					g.Id("Future").Op(":").Qual(workflowPkg, "ExecuteChildWorkflow").CallFunc(func(args *g.Group) {
						args.Id("ctx")
						args.Id(fmt.Sprintf("%sWorkflowName", workflow))
						if hasInput {
							args.Id("req")
						} else {
							args.Nil()
						}
					}),
				),
				g.Nil(),
			)
		})
}

// genWorkerWorkflowResourcesChildWorkflowMethod generates <Workflow>Child method
func (svc *Service) genWorkerWorkflowResourcesChildWorkflowMethod(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	name := strcase.ToCamel(workflow + "Child")
	asyncName := name + "Async"

	f.Commentf("%s executes a child %s workflow", name, workflow)
	f.Func().
		Params(g.Id("r").Op("*").Id(svc.Names().WorkflowResources())).
		Id(name).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual(workflowPkg, "Context")
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			args.Id("options").Op("...").Op("*").Qual(workflowPkg, "ChildWorkflowOptions")
		}).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		BlockFunc(func(fn *g.Group) {
			fn.List(g.Id("childRun"), g.Err()).Op(":=").Id("r").Dot(asyncName).CallFunc(func(args *g.Group) {
				args.Id("ctx")
				if hasInput {
					args.Id("req")
				}
				args.Id("options").Op("...")
			})
			fn.If(g.Err().Op("!=").Nil()).Block(
				g.ReturnFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Nil()
					}
					returnVals.Err()
				}),
			)
			fn.Return(
				g.Id("childRun").Dot("Get").Call(g.Id("ctx")),
			)
		})
}

// genWorkerWorkflowResourcesSignalExternal generates a <Signal>External method
func (svc *Service) genWorkerWorkflowResourcesSignalExternal(f *g.File, signal string) {
	method := svc.methods[signal]
	methodName := svc.Names().SignalGoName(signal)
	signalName := svc.Names().SignalName(signal)
	hasInput := !isEmpty(method.Input)
	f.Commentf("%sExternal sends a %s signal to an existing workflow", methodName, signalName)
	f.Func().
		Params(g.Id("r").Op("*").Id(svc.Names().WorkflowResources())).
		Id(methodName).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual(workflowPkg, "Context")
			args.Id("workflowID").String()
			args.Id("runID").String()
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
		}).
		Params(g.Error()).
		Block(
			g.Return(
				g.Qual(workflowPkg, "SignalExternalWorkflow").CallFunc(func(args *g.Group) {
					args.Id("ctx")
					args.Id("workflowID")
					args.Id("runID")
					args.Id(signalName)
					if hasInput {
						args.Id("req")
					} else {
						args.Nil()
					}
				}).Dot("Get").Call(g.Id("ctx"), g.Nil()),
			),
		)
}

// genWorkerWorkflowsInterface generates a Workflows interface for a given service
func (svc *Service) genWorkerWorkflowsInterface(f *g.File) {
	// generate workflows interface
	f.Commentf("Workflows provides methods for initializing new %s workflow values", svc.GoName)
	f.Type().Id("Workflows").InterfaceFunc(func(methods *g.Group) {
		for _, workflow := range svc.workflowsOrdered {
			// method := svc.methods[workflow]
			methods.Commentf("%s initializes a new %sWorkflow value", workflow, workflow).Line().
				Id(workflow).
				Params(
					g.Id("ctx").Qual(workflowPkg, "Context"),
					g.Id("input").Op("*").Id(fmt.Sprintf("%sInput", workflow)),
				).
				Params(
					g.Id(fmt.Sprintf("%sWorkflow", workflow)),
					g.Error(),
				)
		}
	})
}
