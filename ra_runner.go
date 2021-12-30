package FabricEmu

import (
	"crypto/rsa"
	"embed"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed servicepkgtmpl/*.tpl
var tmpls embed.FS // this folder contains very basic files to run sf app

type serviceManifest struct {
	XMLName      xml.Name `xml:"ServiceManifest"`
	ServiceTypes struct {
		XMLName             xml.Name `xml:"ServiceTypes"`
		StatefulServiceType []struct {
			XMLName         xml.Name `xml:"StatefulServiceType"`
			ServiceTypeName string   `xml:"ServiceTypeName,attr"`
		}
	}
	CodePackage struct {
		XMLName    xml.Name `xml:"CodePackage"`
		Name       string   `xml:"Name,attr"`
		EntryPoint struct {
			XMLName xml.Name `xml:"EntryPoint"`
			ExeHost struct {
				XMLName       xml.Name `xml:"ExeHost"`
				Program       string   `xml:"Program"`
				Arguments     string   `xml:"Arguments"`
				WorkingFolder string   `xml:"WorkingFolder"`
			}
		}
	}
}

func loadServiceManifest(servicePkgPath string) (*serviceManifest, error) {
	manifestfile, err := ioutil.ReadFile(filepath.Join(servicePkgPath, "ServiceManifest.xml"))
	if err != nil {
		return nil, err
	}

	var manifestxml serviceManifest
	if err := xml.Unmarshal(manifestfile, &manifestxml); err != nil {
		return nil, err
	}

	return &manifestxml, nil
}

func (h *ReplicaAgent) buildRuntimeFolder(manifestxml *serviceManifest) error {

	appname := "Dummy" // TODO dup code
	appid := fmt.Sprintf("%v_App0", appname)
	if err := os.MkdirAll(filepath.Join(h.deploymentDirectory, appid), 0755); err != nil {
		return err
	}

	files, err := tmpls.ReadDir("servicepkgtmpl")
	if err != nil {
		return err
	}

	for _, f := range files {
		tf, err := os.Create(filepath.Join(h.deploymentDirectory, appid, strings.TrimSuffix(f.Name(), ".tpl")))
		if err != nil {
			return err
		}

		t, err := template.ParseFS(tmpls, path.Join("servicepkgtmpl", f.Name()))
		if err != nil {
			return err
		}

		if err := t.Execute(tf, manifestxml); err != nil {
			return nil
		}

		tf.Close()
	}

	return nil
}

func (h *ReplicaAgent) ExecServicePkg(servicePkgPath string) (*exec.Cmd, error) {

	manifestxml, err := loadServiceManifest(servicePkgPath)
	if err != nil {
		return nil, err
	}

	if err := h.buildRuntimeFolder(manifestxml); err != nil {
		return nil, err
	}

	// _, appname := filepath.Split(servicePkgPath)

	appname := "Dummy"
	appid := fmt.Sprintf("%v_App0", appname)

	logdir, err := os.MkdirTemp("", "sfemulog")
	if err != nil {
		return nil, err
	}

	clientKeystr, err := cryptExportKey(h.clientCert.PrivateKey.(*rsa.PrivateKey))
	if err != nil {
		return nil, err
	}

	// code here in sf means bin
	codedir := filepath.Join(servicePkgPath, manifestxml.CodePackage.Name)
	cmd := exec.Command(manifestxml.CodePackage.EntryPoint.ExeHost.Program, manifestxml.CodePackage.EntryPoint.ExeHost.Arguments)

	switch manifestxml.CodePackage.EntryPoint.ExeHost.WorkingFolder {
	case "", "CodeBase":
		cmd.Dir = codedir
	case "CodePackage":
		cmd.Dir = servicePkgPath
	default:
		cmd.Dir = manifestxml.CodePackage.EntryPoint.ExeHost.WorkingFolder
	}

	cmd.Env = append(
		os.Environ(),

		fmt.Sprintf("FabricLogRoot=%v", logdir),
		fmt.Sprintf("Fabric_ApplicationHostId=%v", h.hostid),
		"Fabric_ApplicationHostType=Activated_SingleCodePackage", // TODO support other type
		"Fabric_IsContainerHost=true",
		"Fabric_IsCodePackageActivatorHost=false",

		"Fabric_ServicePackageInstanceSeqNum=1", // TODO support sync
		"Fabric_CodePackageInstanceSeqNum=1",
		fmt.Sprintf("Fabric_ApplicationName=%v", appname),
		fmt.Sprintf("Fabric_ApplicationId=%v", appid),

		"Fabric_CodePackageName=Code",
		"Fabric_ServicePackageVersionInstance=1.0:1.0:1",
		"Fabric_ServicePackageName=ServicePkg",

		"Fabric_NetworkingMode=Open;",
		fmt.Sprintf("Fabric_RuntimeSslConnectionAddress=%v", h.listenaddr),
		fmt.Sprintf("Fabric_RuntimeSslConnectionCertThumbprint=%v", h.serverCertTp),

		// this is windows only, sf uses CryptExportKey with PRIVATEKEYBLOB to export keys
		fmt.Sprintf("Fabric_RuntimeSslConnectionCertKey=%v", clientKeystr),
		fmt.Sprintf("Fabric_RuntimeSslConnectionCertEncodedBytes=%v", base64.StdEncoding.EncodeToString(h.clientCert.Certificate[0])),

		// TODO RuntimeSslConnectionCertFilePath on linux
	)

	return cmd, nil
}
