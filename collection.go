package main

import "google.golang.org/protobuf/compiler/protogen"

func generateBrunoCollection(gen *protogen.Plugin, file *protogen.File, prefix string) error {
	for _, service := range file.Services {
		for _, method := range service.Methods {
			if mode == modeAll || mode == modeHTTP {
				if err := generateBrunoRequest(gen, service, method, prefix); err != nil {
					return err
				}
			}
			if mode == modeAll || mode == modeGRPC {
				if err := generateGrpcRequest(gen, service, method, file, prefix); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
