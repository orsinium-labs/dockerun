package dockerun

var Presets = map[string]func() Builder{
	"debian": presetDebian,
	"golang": presetGolang,
	"python": presetPython,
	"ubuntu": presetUbuntu,
}

func baseBuilder() Builder {
	return Builder{
		Prefix:     "dockerun-",
		Name:       "{{name .Package}}",
		EntryPoint: "{{name .Package}}",
		Docker:     &DockerConfig{},
	}
}

func presetDebian() Builder {
	c := baseBuilder()
	c.Image = "debian"
	c.Tag = "stretch"
	c.Install = "apt-get install -y {{.Package}}"
	return c
}

func presetGolang() Builder {
	c := baseBuilder()
	c.Image = "golang"
	c.Tag = "latest"
	c.Install = "go get {{.Package}}"
	return c
}

func presetPython() Builder {
	c := baseBuilder()
	c.Image = "python"
	c.Tag = "3.8"
	c.Install = "python3 -m pip install {{.Package}}"
	return c
}

func presetUbuntu() Builder {
	c := baseBuilder()
	c.Image = "ubuntu"
	c.Tag = "20.04"
	c.Install = "apt-get install -y {{.Package}}"
	return c
}
