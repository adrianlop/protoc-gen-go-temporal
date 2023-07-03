package plugin

import (
	"fmt"

	g "github.com/dave/jennifer/jen"
)

// genTestClientImpl generates a TestClient struct
func (svc *Service) genTestClientImpl(f *g.File) {
	f.Comment("TestClient provides a testsuite-compatible Client")
	f.Type().Id("TestClient").Struct(
		g.Id("env").Op("*").Qual(testsuitePkg, "TestWorkflowEnvironment"),
		g.Id("workflows").Id("Workflows"),
	)
}

// genTestClientImplNewMethod generates a NewTestClient constructor function
func (svc *Service) genTestClientImplNewMethod(f *g.File) {
	f.Var().Id("_").Id("Client").Op("=").Op("&").Id("TestClient").Values()
	f.Comment("NewTestClient initializes a new TestClient value")
	f.Func().Id("NewTestClient").
		Params(
			g.Id("env").Op("*").Qual(testsuitePkg, "TestWorkflowEnvironment"),
			g.Id("workflows").Id("Workflows"),
		).
		Params(g.Op("*").Id("TestClient")).
		Block(
			g.Return(g.Op("&").Id("TestClient").Values(g.Id("env"), g.Id("workflows"))),
		)
}

// genTestClientImplQueryMethod genereates a TestClient <Query> method
func (svc *Service) genTestClientImplQueryMethod(f *g.File, query string) {
	handler := svc.methods[query]
	hasInput := !isEmpty(handler.Input)
	hasOutput := !isEmpty(handler.Output)
	f.Commentf("%s executes a %s query", query, query)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(query).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			args.Id("workflowID").String()
			args.Id("runID").String()
			if hasInput {
				args.Id("req").Op("*").Id(handler.Input.GoIdent.GoName)
			}
		}).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(handler.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		BlockFunc(func(fn *g.Group) {
			fn.List(g.Id("val"), g.Err()).Op(":=").Id("c").Dot("env").Dot("QueryWorkflow").CallFunc(func(args *g.Group) {
				args.Id(fmt.Sprintf("%sQueryName", query))
				if hasInput {
					args.Id("req")
				}
			})
			fn.If(g.Err().Op("!=").Nil()).Block(
				g.ReturnFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Nil()
					}
					returnVals.Err()
				}),
			).Else().If(g.Op("!").Id("val").Dot("HasValue").Call()).Block(
				g.ReturnFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Nil()
					}
					returnVals.Nil()
				}),
			).Else().BlockFunc(func(bl *g.Group) {
				if !hasOutput {
					bl.Return(g.Nil())
				} else {
					bl.Var().Id("result").Id(handler.Output.GoIdent.GoName)
					bl.If(g.Err().Op(":=").Id("val").Dot("Get").Call(g.Op("&").Id("result")), g.Err().Op("!=").Nil()).Block(
						g.Return(
							g.Nil(),
							g.Err(),
						),
					)
					bl.Return(g.Op("&").Id("result"), g.Nil())
				}
			})
		})
}

// genTestClientImplSignalMethod genereates a TestClient <Signal> method
func (svc *Service) genTestClientImplSignalMethod(f *g.File, signal string) {
	handler := svc.methods[signal]
	hasInput := !isEmpty(handler.Input)
	f.Commentf("%s executes a %s signal", signal, signal)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(signal).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			args.Id("workflowID").String()
			args.Id("runID").String()
			if hasInput {
				args.Id("req").Op("*").Id(handler.Input.GoIdent.GoName)
			}
		}).
		Params(
			g.Error(),
		).
		Block(
			g.Id("c").Dot("env").Dot("SignalWorkflow").CallFunc(func(args *g.Group) {
				args.Id(fmt.Sprintf("%sSignalName", signal))
				if hasInput {
					args.Id("req")
				} else {
					args.Nil()
				}
			}),
			g.Return(g.Nil()),
		)
}

// genTestClientImplUpdateMethod genereates a TestClient <Update> method
func (svc *Service) genTestClientImplUpdateMethod(f *g.File, update string) {

}

// genTestClientImplUpdateAsyncMethod genereates a TestClient <UpdateAsync> method
func (svc *Service) genTestClientImplUpdateAsyncMethod(f *g.File, update string) {

}

// genTestClientImplWorkflowMethod generates a TestClient <Workflow> method
func (svc *Service) genTestClientImplWorkflowMethod(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	f.Commentf("%s executes a(n) %s workflow in the test environment", workflow, workflow)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(workflow).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			args.Id("opts").Op("...").Op("*").Qual(clientPkg, "StartWorkflowOptions")
		}).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		Block(
			g.List(g.Id("run"), g.Err()).Op(":=").Id("c").Dot(fmt.Sprintf("%sAsync", workflow)).CallFunc(func(args *g.Group) {
				args.Id("ctx")
				if hasInput {
					args.Id("req")
				}
				args.Id("opts").Op("...")
			}),
			g.If(g.Err().Op("!=").Nil()).Block(
				g.ReturnFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Nil()
					}
					returnVals.Err()
				}),
			),
			g.Return(g.Id("run").Dot("Get").Call(g.Id("ctx"))),
		)
}

// genTestClientImplWorkflowAsyncMethod generates a TestClient's <workflow>Async method
func (svc *Service) genTestClientImplWorkflowAsyncMethod(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	f.Commentf("%sAsync executes a(n) %s workflow in the test environment", workflow, workflow)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(fmt.Sprintf("%sAsync", workflow)).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			args.Id("options").Op("...").Op("*").Qual(clientPkg, "StartWorkflowOptions")
		}).
		Params(
			g.Id(fmt.Sprintf("%sRun", workflow)),
			g.Error(),
		).
		BlockFunc(func(fn *g.Group) {
			svc.genClientStartWorkflowOptions(fn, workflow, false)
			fn.Return(
				g.Op("&").Id(fmt.Sprintf("test%sRun", workflow)).ValuesFunc(func(fields *g.Group) {
					fields.Id("env").Op(":").Id("c").Dot("env")
					fields.Id("opts").Op(":").Id("opts")
					if hasInput {
						fields.Id("req").Op(":").Id("req")
					}
					fields.Id("workflows").Op(":").Id("c").Dot("workflows")
				}),
				g.Nil(),
			)
		})
}

// genTestClientImplWorkflowGetMethod generates a TestClient's Get<workflow> method
func (svc *Service) genTestClientImplWorkflowGetMethod(f *g.File, workflow string) {
	f.Commentf("Get%s is a noop", workflow)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(fmt.Sprintf("Get%s", workflow)).
		Params(
			g.Id("ctx").Qual("context", "Context"),
			g.Id("workflowID").String(),
			g.Id("runID").String(),
		).
		Params(
			g.Id(fmt.Sprintf("%sRun", workflow)),
			g.Error(),
		).
		Block(
			g.Return(
				g.Op("&").Id(fmt.Sprintf("test%sRun", workflow)).Values(
					g.Id("env").Op(":").Id("c").Dot("env"),
					g.Id("workflows").Op(":").Id("c").Dot("workflows"),
				),
				g.Nil(),
			),
		)
}

// genTestClientImplWorkflowWithSignalMethod generates a TestClient's <workflow>With<signal> method
func (svc *Service) genTestClientImplWorkflowWithSignalMethod(f *g.File, workflow, signal string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	handler := svc.methods[signal]
	hasSignalInput := !isEmpty(handler.Input)
	f.Commentf("%sWith%s sends a(n) %s signal to a(n) %s workflow, starting it if necessary", workflow, signal, signal, workflow)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(fmt.Sprintf("%sWith%s", workflow, signal)).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			if hasSignalInput {
				args.Id("signal").Op("*").Id(handler.Input.GoIdent.GoName)
			}
			args.Id("opts").Op("...").Op("*").Qual(clientPkg, "StartWorkflowOptions")
		}).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		Block(
			g.Id("c").Dot("env").Dot("RegisterDelayedCallback").Call(
				g.Func().Params().Block(
					g.Id("c").Dot("env").Dot("SignalWorkflow").CallFunc(func(args *g.Group) {
						args.Id(fmt.Sprintf("%sSignalName", signal))
						if hasSignalInput {
							args.Id("signal")
						} else {
							args.Nil()
						}
					}),
				),
				g.Lit(0),
			),
			g.Return(
				g.Id("c").Dot(workflow).CallFunc(func(args *g.Group) {
					args.Id("ctx")
					if hasInput {
						args.Id("req")
					}
					args.Id("opts").Op("...")
				}),
			),
		)
}

// genTestClientImplWorkflowWithSignalAsyncMethod generates a TestClient's <workflow>With<signal>Async method
func (svc *Service) genTestClientImplWorkflowWithSignalAsyncMethod(f *g.File, workflow, signal string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	handler := svc.methods[signal]
	hasSignalInput := !isEmpty(handler.Input)
	f.Commentf("%sWith%sAsync sends a(n) %s signal to a(n) %s workflow, starting it if necessary", workflow, signal, signal, workflow)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Id(fmt.Sprintf("%sWith%sAsync", workflow, signal)).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			if hasInput {
				args.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
			}
			if hasSignalInput {
				args.Id("signal").Op("*").Id(handler.Input.GoIdent.GoName)
			}
			args.Id("opts").Op("...").Op("*").Qual(clientPkg, "StartWorkflowOptions")
		}).
		Params(
			g.Op("*").Id(fmt.Sprintf("test%sRun", workflow)),
			g.Error(),
		).
		Block(
			g.Id("c").Dot("env").Dot("RegisterDelayedCallback").Call(
				g.Func().Params().Block(
					g.Id("_").Op("=").Id("c").Dot(signal).CallFunc(func(args *g.Group) {
						args.Id("ctx")
						args.Lit("")
						args.Lit("")
						if hasInput {
							args.Id("signal")
						}
					}),
				),
				g.Lit(0),
			),
			g.Return(
				g.Id("c").Dot(fmt.Sprintf("%sAsync", workflow)).CallFunc(func(args *g.Group) {
					args.Id("ctx")
					if hasInput {
						args.Id("req")
					}
					args.Id("opts").Op("...")
				}),
			),
		)
}

// genTestClientWorkflowGetMethod generates a noop TestClient Get<Workflow> method
func (svc *Service) genTestClientWorkflowGetMethod(f *g.File, workflow string) {
	f.Commentf("Get%s retrieves a test %sRun", workflow, workflow)
	f.Func().
		Params(g.Id("c").Op("*").Id("TestClient")).
		Params(
			g.Qual("context", "Context"),
			g.String(),
			g.String(),
		).
		Params(
			g.Id(fmt.Sprintf("%sRun", workflow)),
			g.Error(),
		).
		Block(
			g.Return(g.Op("&").Id(fmt.Sprintf("test%sRun", workflow)).Values(g.Id("env").Op(":").Id("c").Dot("env"))),
		)
}

// genTestClientWorkflowRunImpl generates a test<Workflow>Run struct
func (svc *Service) genTestClientWorkflowRunImpl(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	// generate test<Workflow>Run struct
	f.Var().Id("_").Id(fmt.Sprintf("%sRun", workflow)).Op("=").Op("&").Id(fmt.Sprintf("test%sRun", workflow)).Values()
	f.Commentf("test%sRun provides convenience methods for interacting with a(n) %s workflow in the test environment", workflow, workflow)
	f.Type().Id(fmt.Sprintf("test%sRun", workflow)).StructFunc(func(fields *g.Group) {
		fields.Id("env").Op("*").Qual(testsuitePkg, "TestWorkflowEnvironment")
		fields.Id("opts").Op("*").Qual(clientPkg, "StartWorkflowOptions")
		if hasInput {
			fields.Id("req").Op("*").Id(method.Input.GoIdent.GoName)
		}
		fields.Id("workflows").Id("Workflows")
	})
}

// genTestClientWorkflowRunImplGetMethod generates a test<Workflow>Run's Get method
func (svc *Service) genTestClientWorkflowRunImplGetMethod(f *g.File, workflow string) {
	method := svc.methods[workflow]
	hasInput := !isEmpty(method.Input)
	hasOutput := !isEmpty(method.Output)
	f.Commentf("Get retrieves a test %s workflow result", workflow)
	f.Func().
		Params(g.Id("r").Op("*").Id(fmt.Sprintf("test%sRun", workflow))).
		Id("Get").
		Params(g.Qual("context", "Context")).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(method.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).
		BlockFunc(func(fn *g.Group) {
			// execute workflow
			fn.Id("r").Dot("env").Dot("ExecuteWorkflow").CallFunc(func(args *g.Group) {
				args.Id(fmt.Sprintf("build%s", workflow)).Call(g.Id("r").Dot("workflows").Dot(workflow))
				if hasInput {
					args.Id("r").Dot("req")
				}
			})
			// ensure completed
			fn.If(g.Op("!").Id("r").Dot("env").Dot("IsWorkflowCompleted").Call()).Block(
				g.ReturnFunc(func(returnVals *g.Group) {
					if hasOutput {
						returnVals.Nil()
					}
					returnVals.Qual("errors", "New").Call(g.Lit("workflow in progress"))
				}),
			)
			if hasOutput {
				fn.Var().Id("result").Id(method.Output.GoIdent.GoName)
				fn.If(g.Err().Op(":=").Id("r").Dot("env").Dot("GetWorkflowResult").Call(g.Op("&").Id("result")), g.Err().Op("!=").Nil()).Block(
					g.Return(g.Nil(), g.Err()),
				)
				fn.Return(g.Op("&").Id("result"), g.Nil())
			} else {
				fn.Return(g.Nil())
			}
		})
}

// genTestClientWorkflowRunImplIDMethod generates a test<Workflow>Run's workflow ID
func (svc *Service) genTestClientWorkflowRunImplIDMethod(f *g.File, workflow string) {
	f.Commentf("ID returns a test %s workflow run's workflow ID", workflow)
	f.Func().
		Params(g.Id("r").Op("*").Id(fmt.Sprintf("test%sRun", workflow))).
		Id("ID").
		Params().
		Params(g.String()).
		Block(
			g.If(g.Id("r").Dot("opts").Op("!=").Nil()).Block(
				g.Return(g.Id("r").Dot("opts").Dot("ID")),
			),
			g.Return(g.Lit("")),
		)
}

// genTestClientWorkflowRunImplQueryMethod generates a test<Workflow>Run's <Query> method
func (svc *Service) genTestClientWorkflowRunImplQueryMethod(f *g.File, workflow, query string) {
	handler := svc.methods[query]
	hasInput := !isEmpty(handler.Input)
	hasOutput := !isEmpty(handler.Output)
	f.Commentf("%s executes a %s query against a test %s workflow", query, query, workflow)
	f.Func().
		Params(g.Id("r").Op("*").Id(fmt.Sprintf("test%sRun", workflow))).
		Id(query).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			if hasInput {
				args.Id("req").Op("*").Id(handler.Input.GoIdent.GoName)
			}
		}).
		ParamsFunc(func(returnVals *g.Group) {
			if hasOutput {
				returnVals.Op("*").Id(handler.Output.GoIdent.GoName)
			}
			returnVals.Error()
		}).Block(
		g.Return(
			g.Id("c").Dot(query).CallFunc(func(args *g.Group) {
				args.Id("ctx")
				args.Lit("")
				args.Lit("")
				if hasInput {
					args.Id("req")
				}
			}),
		),
	)
}

// genTestClientWorkflowRunImplRunIDMethod generates a test<Workflow>Run's RunID method
func (svc *Service) genTestClientWorkflowRunImplRunIDMethod(f *g.File, workflow string) {
	f.Comment("RunID noop implementation")
	f.Func().
		Params(g.Id("r").Op("*").Id(fmt.Sprintf("test%sRun", workflow))).
		Id("RunID").
		Params().
		Params(g.String()).
		Block(
			g.Return(g.Lit("")),
		)
}

// genTestClientWorkflowRunImplQueryMethod generates a test<Workflow>Run's <Signal> method
func (svc *Service) genTestClientWorkflowRunImplSignalMethod(f *g.File, workflow, signal string) {
	handler := svc.methods[signal]
	hasInput := !isEmpty(handler.Input)
	f.Commentf("%s executes a %s signal against a test %s workflow", signal, signal, workflow)
	f.Func().
		Params(g.Id("r").Op("*").Id(fmt.Sprintf("test%sRun", workflow))).
		Id(signal).
		ParamsFunc(func(args *g.Group) {
			args.Id("ctx").Qual("context", "Context")
			if hasInput {
				args.Id("req").Op("*").Id(handler.Input.GoIdent.GoName)
			}
		}).
		Params(
			g.Error(),
		).
		Block(
			g.Return(
				g.Id("c").Dot(signal).CallFunc(func(args *g.Group) {
					args.Id("ctx")
					args.Lit("")
					args.Lit("")
					if hasInput {
						args.Id("req")
					}
				}),
			),
		)
}

// genTestClientWorkflowRunImplQueryMethod generates a test<Workflow>Run's <Update> method
func (svc *Service) genTestClientWorkflowRunImplUpdateMethod(f *g.File, workflow, update string) {

}

// genTestClientWorkflowRunImplQueryMethod generates a test<Workflow>Run's <Update>Async method
func (svc *Service) genTestClientWorkflowRunImplUpdateAsyncMethod(f *g.File, workflow, update string) {

}
