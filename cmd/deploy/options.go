package deploy

import (
	"fmt"
	"os"
	"github.com/spf13/pflag"

	"github.com/ydb-platform/ydbops/pkg/cmdutil"
	"gopkg.in/yaml.v2"
	// "github.com/ydb-platform/ydbops/pkg/prettyprint"
)

type Options struct {
	Config string
	Settings string
}

func (o *Options) DefineFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Config, "config", "",
		`YDB cluster YAML config`)
	fs.StringVar(&o.Settings, "settings", "",
		"YAML with settings (ansible inventory)")
}

func (o *Options) Validate() error {
	// TODO(shmel1k@): remove copypaste between drop, create & refresh methods.
	if o.Config == "" {
		return fmt.Errorf("--config unspecified")
	}
	if o.Settings == "" {
		return fmt.Errorf("--settings unspecified, argument required")
	}
	return nil
}

func (o *Options) Run(f cmdutil.Factory) error {
	// file, err := os.Open(o.Config.FilePath)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	var config map[string]interface{}
	data, _ := os.ReadFile(o.Config)

	if err := yaml.Unmarshal(data, &config); err != nil {
    	return err
  	}

	fmt.Printf("Parsed config: %+v\n", config["hosts"])

	hosts,ok := config["hosts"].(map[string]interface{})
	if !ok {
    	panic("hosts is not a map")
  	}

	fmt.Printf("%+v",hosts)

	// for version, settings := range hosts {
	// 	fmt.Printf("%s %s",version, settings)
	// }


	// result, err := f.GetCMSClient().CompleteActions(o.Config, o.Settings)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(prettyprint.ResultToString(result))
	return nil
}
