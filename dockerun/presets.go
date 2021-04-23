package dockerun

var Presets = map[string]func() Builder{
	"go":     presetGo,
	"python": presetPython,
}

func baseBuilder() Builder {
	return Builder{
		Prefix:     "_dockerun",
		Name:       "{{.Prefix}}-{{.Package}}",
		EntryPoint: "bash -c {{.Package}}",
		Docker:     &DockerConfig{},
	}
}

func presetPython() Builder {
	c := baseBuilder()
	c.Image = "python"
	c.Tag = "3.8"
	c.Install = "python3 -m pip install {{.Package}}"
	return c
}

func presetGo() Builder {
	c := baseBuilder()
	c.Image = "golang"
	c.Tag = "latest"
	c.Install = "go get {{.Package}}"
	return c
}
