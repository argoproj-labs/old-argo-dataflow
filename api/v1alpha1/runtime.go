package v1alpha1

type Runtime string

func (r Runtime) HandlerFile() string {
	switch r {
	case "go1.16":
		return "handler.go"
	case "java16":
		return "Handler.java"
	case "python3-9":
		return "handler.py"
	default:
		panic(r)
	}
}

func (r Runtime) GetImage() string {
	switch r {
	case "go1.16":
		return "golang:1.16"
	case "java16":
		return "openjdk:16"
	case "python3-9":
		return "python:3.9"
	default:
		panic(r)
	}
}
