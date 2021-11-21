//
// Copyright © 2021 Kris Nóva <kris@nivenly.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//   ███╗   ██╗ █████╗ ███╗   ███╗██╗
//   ████╗  ██║██╔══██╗████╗ ████║██║
//   ██╔██╗ ██║███████║██╔████╔██║██║
//   ██║╚██╗██║██╔══██║██║╚██╔╝██║██║
//   ██║ ╚████║██║  ██║██║ ╚═╝ ██║███████╗
//   ╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝
//

package codify

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/kris-nova/logger"
	corev1 "k8s.io/api/core/v1"
)

type PersistentVolumeClaim struct {
	KubeObject *corev1.PersistentVolumeClaim
	GoName     string
}

func NewPersistentVolumeClaim(obj *corev1.PersistentVolumeClaim) *PersistentVolumeClaim {
	obj.ObjectMeta = cleanObjectMeta(obj.ObjectMeta)
	obj.Status = corev1.PersistentVolumeClaimStatus{}
	return &PersistentVolumeClaim{
		KubeObject: obj,
		GoName:     goName(obj.Name),
	}
}

func (k PersistentVolumeClaim) Install() (string, []string) {

	// Ignore resource requirements
	k.KubeObject.Spec.Resources = corev1.ResourceRequirements{}

	c, err := Literal(k.KubeObject)
	if err != nil {
		logger.Debug(err.Error())
	}
	l := c.Source
	packages := c.Packages
	install := fmt.Sprintf(`
	{{ .GoName }}PersistentVolumeClaim := %s
	x.objects = append(x.objects, {{ .GoName }}PersistentVolumeClaim)

	if client != nil {
		_, err = client.CoreV1().PersistentVolumeClaims("{{ .KubeObject.Namespace }}").Create(context.TODO(), {{ .GoName }}PersistentVolumeClaim, v1.CreateOptions{})
		if err != nil {
			return err
		}
	}
`, l)

	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl.Parse(install)
	buf := &bytes.Buffer{}
	k.KubeObject.Name = sanitizeK8sObjectName(k.KubeObject.Name)
	err = tpl.Execute(buf, k)
	if err != nil {
		logger.Debug(err.Error())
	}
	return alias(buf.String(), "corev1"), packages
}

func (k PersistentVolumeClaim) Uninstall() string {
	uninstall := `
	if client != nil {
		err = client.CoreV1().PersistentVolumeClaims("{{ .KubeObject.Namespace }}").Delete(context.TODO(), "{{ .KubeObject.Name }}", metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
 `
	tpl := template.New(fmt.Sprintf("%s", time.Now().String()))
	tpl.Parse(uninstall)
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, k)
	if err != nil {
		logger.Debug(err.Error())
	}
	return buf.String()
}
