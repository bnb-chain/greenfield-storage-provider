package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestSaveConfigParser(t *testing.T) {
	f, err := os.Create("./gfsp_config.toml")
	fmt.Println(err)
	require.NoError(t, err)
	defer f.Close()
	cfg := &GfSpConfig{
		AppID:  "asdasd",
		GfSpDB: nil,
	}
	b, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	f.WriteString(string(b))
	//encode := util.TomlSettings.NewEncoder(f)
	//err = encode.Encode(cfg)
	//fmt.Println(err)
	require.NoError(t, err)

	bz, _ := ioutil.ReadFile("./gfsp_config.toml")
	cfg1 := &GfSpConfig{}
	toml.Unmarshal(bz, cfg1)
	fmt.Printf("%v", cfg1)
}
