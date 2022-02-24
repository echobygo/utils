package YamlRead
import (
	"errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path"
)


// 将yaml文件中的内容进行加载
func Load (path string ,result interface{}) error {
	ext := guessFileType(path)
	if !ext  {
		return errors.New("cant not load" + path + " config")
	}
	return loadFromYaml(path,result)
}

//判断配置文件名是否为yaml格式
func  guessFileType(pathFile string) bool {
	 resultFileName:=path.Base(pathFile)
	 if path.Ext(resultFileName)==".yaml"{
	 	return true;
	 }
	 return false;
	//s := strings.Split(path,".")
	//ext := s[len(s) - 1]
	//switch ext {
	//case "yaml","yml":
	//	return "yaml"
	//}
	//return ""
}

// 将配置yaml文件中的进行加载
func  loadFromYaml(path string ,result interface{}) error {
	yamlS,readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	// yaml解析的时候c.data如果没有被初始化，会自动为你做初始化
	err := yaml.Unmarshal(yamlS, result)
	if err != nil {
		return errors.New("can not parse "+ path + " config" +err.Error())
	}
	return nil
}
