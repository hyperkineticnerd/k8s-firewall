package juniper

import (
	"github.com/Juniper/go-netconf/netconf"
	"github.com/hyperkineticnerd/k8s-firewall/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type JuniperConfig struct {
	Host string
}

type JuniperConnection struct {
	Config  *ssh.ClientConfig
	Session *netconf.Session
}

func BuildConfig(cfg *config.Config) *ssh.ClientConfig {
	var config *ssh.ClientConfig

	var sshConfig ssh.Config
	sshConfig.SetDefaults()
	cipherOrder := sshConfig.Ciphers
	sshConfig.Ciphers = append(cipherOrder, "aes128-cbc")

	logrus.Debugf("Juniper Config")
	config = &ssh.ClientConfig{
		Config:          sshConfig,
		User:            cfg.SSHUsername,
		Auth:            []ssh.AuthMethod{ssh.Password(cfg.SSHPassphrase)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config
}

func JuniperSetup(cfg *config.Config) *JuniperConnection {
	config := BuildConfig(cfg)
	ssh, err := netconf.DialSSH(cfg.RouterHost, config)
	logrus.Debugf("Juniper Connecting...")
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Debugf("Juniper Connected ???")
	return &JuniperConnection{
		Config:  config,
		Session: ssh,
	}
}

func (j *JuniperConnection) TextMethodEditConfig(txtData string) netconf.RawMethod {
	logrus.Debugf("TextMethodEditConfig")
	return netconf.RawMethod(txtData)
}

func (j *JuniperConnection) Edit(data string) {
	jdata := j.TextMethodEditConfig(data)
	logrus.Debugf("jdata: %+v", jdata)

	reply, err := j.Session.Exec(jdata)
	if err != nil {
		logrus.Errorf("Edit %v", err)
	}
	logrus.Debugf("Reply: %+v", reply)
}
