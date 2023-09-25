package FabricEmu

import (
	"crypto/rsa"
	"embed"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:servicepkgtmpl
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

func (h *ReplicaAgent) buildRuntimeFolderTpl(basedir, subdir string, manifestxml *serviceManifest) error {
	files, err := tmpls.ReadDir(path.Join("servicepkgtmpl", subdir))
	if err != nil {
		return err
	}

	for _, f := range files {

		if f.IsDir() {
			if err := os.MkdirAll(filepath.Join(basedir, subdir, f.Name()), 0755); err != nil {
				return err
			}

			if err := h.buildRuntimeFolderTpl(basedir, path.Join(subdir, f.Name()), manifestxml); err != nil {
				return err
			}

			continue
		}

		tf, err := os.Create(filepath.Join(basedir, subdir, strings.TrimSuffix(f.Name(), ".tpl")))
		if err != nil {
			return err
		}

		if strings.HasSuffix(f.Name(), ".tpl") {
			t, err := template.ParseFS(tmpls, path.Join("servicepkgtmpl", subdir, f.Name()))
			if err != nil {
				return err
			}

			if err := t.Execute(tf, manifestxml); err != nil {
				return nil
			}
		} else {

			b, err := tmpls.ReadFile(path.Join("servicepkgtmpl", subdir, f.Name()))
			if err != nil {
				return err
			}

			if _, err := tf.Write(b); err != nil {
				return err
			}
		}

		tf.Close()
	}

	return nil
}

func (h *ReplicaAgent) buildRuntimeFolder(manifestxml *serviceManifest, configDir string) error {

	appname := "Dummy" // TODO dup code
	appid := fmt.Sprintf("%v_App0", appname)
	basedir := filepath.Join(h.deploymentDirectory, appid)
	if err := os.MkdirAll(basedir, 0755); err != nil {
		return err
	}

	// files, err := tmpls.ReadDir("servicepkgtmpl")
	// if err != nil {
	// 	return err
	// }

	if err := h.buildRuntimeFolderTpl(basedir, "", manifestxml); err != nil {
		return err
	}

	if err := copyDir(configDir, filepath.Join(basedir, "ServicePkg.Config.1.0")); err != nil {
		return err
	}

	return nil
}

// chatgpt
func copyDir(sourceDir, destinationDir string) error {
	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destinationDir, 0755); err != nil {
		return err
	}

	// Open the source directory
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		return err
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", sourceDir)
	}

	// Read the contents of the source directory
	fileList, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	// Copy each file and subdirectory from the source to the destination
	for _, file := range fileList {
		sourcePath := filepath.Join(sourceDir, file.Name())
		destinationPath := filepath.Join(destinationDir, file.Name())

		if file.IsDir() {
			// Recursively copy subdirectories
			if err := copyDir(sourcePath, destinationPath); err != nil {
				return err
			}
		} else {
			// Copy files
			sourceFile, err := os.Open(sourcePath)
			if err != nil {
				return err
			}
			defer sourceFile.Close()

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				return err
			}
			defer destinationFile.Close()

			_, err = io.Copy(destinationFile, sourceFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *ReplicaAgent) ExecServicePkg(servicePkgPath, exePath, sfPath, configDir string) (*exec.Cmd, error) {

	manifestxml, err := loadServiceManifest(servicePkgPath)
	if err != nil {
		return nil, err
	}

	if err := h.buildRuntimeFolder(manifestxml, configDir); err != nil {
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
	// codedir := filepath.Join(servicePkgPath, manifestxml.CodePackage.Name)
	codedir := filepath.Join(servicePkgPath, manifestxml.CodePackage.Name)
	// cmd := exec.Command(manifestxml.CodePackage.EntryPoint.ExeHost.Program, manifestxml.CodePackage.EntryPoint.ExeHost.Arguments)
	cmd := exec.Command(exePath, manifestxml.CodePackage.EntryPoint.ExeHost.Arguments)

	switch manifestxml.CodePackage.EntryPoint.ExeHost.WorkingFolder {
	case "", "CodeBase":
		cmd.Dir = codedir
	case "CodePackage":
		cmd.Dir = servicePkgPath
	default:
		cmd.Dir = manifestxml.CodePackage.EntryPoint.ExeHost.WorkingFolder
	}

	// temp workaround
	cmd.Dir = path.Dir(exePath)

	envvar := os.Environ()
	for i, v := range envvar {
		if strings.HasPrefix(v, "Path=") {
			envvar[i] = fmt.Sprintf("Path=%v;%v", sfPath, v[5:])
		}
	}



	cmd.Env = append(
		envvar,

		"AISC_METRIC_DEFAULT_DIMENSION_Region=westus2",
        "AISC_METRIC_DEFAULT_DIMENSION_Environment=test",
        "AISC_METRIC_DEFAULT_DIMENSION_ClusterName=Local",
        "AISC_LOG_NAMESPACE_CUST_WINDOWS=AISCCustTestWindows",
        "AISC_LOG_NAMESPACE_CUST_LINUX=AISCCustTestLinux",
        "ASPNETCORE_ENVIRONMENT=Development.Test",
        "DOTNET_ENVIRONMENT=Development.Test",

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
