package util

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func ReadFile(filename string) (*unstructured.UnstructuredList, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return SplitYAML(data)
}

func SplitYAML(v interface{}) (*unstructured.UnstructuredList, error) {
	switch data := v.(type) {
	case []byte:
		list := &unstructured.UnstructuredList{}
		for _, text := range strings.Split(string(data), "---") {
			item := unstructured.Unstructured{}
			if err := yaml.Unmarshal([]byte(text), &item); err != nil {
				return nil, err
			}
			list.Items = append(list.Items, item)
		}
		return list, nil
	case string:
		return SplitYAML([]byte(data))
	default:
		panic("unknown type")
	}
}

type FileInfoUnstructuredList struct {
	fs.FileInfo
	unstructured.UnstructuredList
}

func ReadDir(dirname string) ([]FileInfoUnstructuredList, error) {
	infos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	sort.Slice(infos, func(i, j int) bool {
		return strings.Compare(infos[i].Name(), infos[j].Name()) < 0
	})
	var x []FileInfoUnstructuredList
	for _, info := range infos {
		if !strings.HasSuffix(info.Name(), ".yaml") {
			continue
		}
		list, err := ReadFile(filepath.Join(dirname, info.Name()))
		if err != nil {
			return nil, err
		}
		x = append(x, FileInfoUnstructuredList{FileInfo: info, UnstructuredList: *list})
	}
	return x, nil
}
