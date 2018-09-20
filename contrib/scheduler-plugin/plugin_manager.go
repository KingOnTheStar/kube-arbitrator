package scheduler_plugin

import (
	"github.com/golang/glog"
	"io/ioutil"
	"strings"

	"k8s.io/api/core/v1"
	"plugin"
)

var schedulerPluginHolder map[v1.ResourceName]SchedulerPlugin

func RegisterDLL(dllpath string) (SchedulerPlugin, error) {
	p := SchedulerPlugin{}
	var dlfunc interface{}
	pdll, err := plugin.Open(dllpath)
	if err != nil {
		glog.Errorf("Open DLL Fail")
		return p, err
	}

	dlfunc, err = pdll.Lookup("HelloDLL")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.HelloDLL = dlfunc.(func(string))

	dlfunc, err = pdll.Lookup("Init")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.Init = dlfunc.(func())

	dlfunc, err = pdll.Lookup("GetResourceName")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.GetResourceName = dlfunc.(func() string)

	dlfunc, err = pdll.Lookup("OnAddNode")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.OnAddNode = dlfunc.(func(string, map[string]string))

	dlfunc, err = pdll.Lookup("OnUpdateNode")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.OnUpdateNode = dlfunc.(func(string, map[string]string))

	dlfunc, err = pdll.Lookup("OnDeleteNode")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.OnDeleteNode = dlfunc.(func(string))

	dlfunc, err = pdll.Lookup("AssessTaskAndNode")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.AssessTaskAndNode = dlfunc.(func(string, int) (int, map[string]string))

	dlfunc, err = pdll.Lookup("OnAddTask")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.OnAddTask = dlfunc.(func(string, map[string]string))

	dlfunc, err = pdll.Lookup("OnRemoveTask")
	if err != nil {
		glog.Errorf("Find DLL Symbol Fail")
		return p, err
	}
	p.OnRemoveTask = dlfunc.(func(string, map[string]string))

	glog.V(3).Infof("SchedulerPlugin init success")

	return p, nil
}

func SearchPlugin(dirpath string) error {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		glog.Errorf("Read Dir Fail, %v", err)
		return err
	}
	// Split to slice and remove empty string
	rawDirpathBuff := strings.Split(dirpath, "/")
	var dirpathBuff []string
	for _, str := range rawDirpathBuff {
		if str != "" {
			dirpathBuff = append(dirpathBuff, str)
		}
	}
	// Init schedulerPluginHolder
	if schedulerPluginHolder == nil {
		schedulerPluginHolder = make(map[v1.ResourceName]SchedulerPlugin)
	}
	// Register each plugin
	for _, file := range files {
		buff := append(dirpathBuff, file.Name())
		dllpath := strings.Join(buff, "/")
		glog.V(3).Infof("Load dll plugin: %s", dllpath)
		schedulerPlugin, err := RegisterDLL(dllpath)
		if err != nil {
			glog.V(3).Infof("Give up dll plugin - %s", dllpath)
			continue
		}
		schedulerPlugin.Init()
		resourceName := v1.ResourceName(schedulerPlugin.GetResourceName())
		schedulerPluginHolder[resourceName] = schedulerPlugin
	}
	return nil
}

func HelloDLL(str string) {
	for _, scheplugin := range schedulerPluginHolder {
		scheplugin.HelloDLL(str)
	}
}

func OnAddNode(node *v1.Node) {
	for _, scheplugin := range schedulerPluginHolder {
		scheplugin.OnAddNode(node.Name, node.Annotations)
	}
}

func OnUpdateNode(node *v1.Node) {
	for _, scheplugin := range schedulerPluginHolder {
		scheplugin.OnUpdateNode(node.Name, node.Annotations)
	}
}

func OnDeleteNode(node *v1.Node) {
	for _, scheplugin := range schedulerPluginHolder {
		scheplugin.OnDeleteNode(node.Name)
	}
}

// TODO
func AssessTaskAndNode(nodeName string, extendDevices map[v1.ResourceName]float64) (int, map[string]string) {
	nodeScore := 0
	nodeAnnotation := make(map[string]string)
	for rName, rValue := range extendDevices {
		scheplugin, ok := schedulerPluginHolder[rName]
		// If this device has no scheduler plugin
		if !ok {
			continue
		}
		requireNum := int(rValue)
		// Call scheduler plugin
		resourceScore, resourceAnnotation := scheplugin.AssessTaskAndNode(nodeName, requireNum)
		// Store in nodeScore and nodeAnnotatoin
		if resourceScore < 0 || nodeScore < 0 {
			nodeScore = -1
		} else {
			nodeScore += resourceScore
			for k, v := range resourceAnnotation {
				nodeAnnotation[k] = v
			}
		}
	}
	return nodeScore, nodeAnnotation
}

func OnAddTask(node *v1.Node, pod *v1.Pod) {
	for _, scheplugin := range schedulerPluginHolder {
		scheplugin.OnAddTask(node.Name, pod.Annotations)
	}
}

func OnRemoveTask(node *v1.Node, pod *v1.Pod) {
	for _, scheplugin := range schedulerPluginHolder {
		scheplugin.OnRemoveTask(node.Name, pod.Annotations)
	}
}
