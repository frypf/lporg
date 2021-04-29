package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/blacktop/lporg/database/utils"
	"github.com/pkg/errors"
)

var porg = `
                                          '.:/+ooossoo+/:-'
                                     ':+ydNMMMMMMMMMMMMMMMNmyo:'
                                   '.--.''.:ohNMMMMMMMMNho:.''..'
                                        -o+.  ':sNMMms-'  .--
                                 '+o     .mNo    '::'   :dNh'    '+-
                                :mMo      dMM-         .NMMs      hNs
                               -NMMNs:--/hMMM+         .MMMNo-''-sMMMs
                          -'   sMMMMMMNNMMMMM:          hMMMMNNNNMMMMN
                         -y    /MMMMMMMMMMMNs           .mMMMMMMMMMMMd     :
                        .mN.    oNMMMMMMMms-             .yNMMMMMMMNy.    -N:
                        hMMm+'   ./syys+-'    -//   ':/-   ./syyys/.    'oNMm'
                       /MMMMMms:.              '.    '''             ./ymMMMMs
                       mMMMMMMMMNo                 '               'sNMMMMMMMN.
                      -MMMMMMMMN+'             :shmmmh+.            'oNMMMMMMMo
                      /MMMMMMMd-             :dds+///sdm/             :mMMMMMMm
                      sMMMMMMd.             :m+'      ':hs'            -mMMMMMM+
                    .hMMMMMMM-             :h:'         'oh.            /MMMMMMN/
                   /mMMMMMMMN'           ./+' .://:::::.  /d/           .MMMMMMMN/
                 'sMMMMMMMMMM/       '.-:-'     '....'     .so-'        oMMMMMMMMN:
                .hMMMMMMMMMMMNs-'     ''                     '--     '-yNMMMMMMMMMN-
               -dMMMMMMMMMMMMMMNs'                                  'yNMMMMMMMMMMMMm.
              -mMMMMMMMMMMMMMMNo'                                    'sMMMMMMMMMMMMMd'
             :NMMMMMMMMMMMMMMm:                                        /NMMMMMMMMMMMMy
            -NMMMMMMMMMMMMMMd.                                        ' -mMMMMMMMMMMMMo
           .mMMMMMMMMMMMMMMd-/o'                                      .o::mMMMMMMMMMMMN/
          'dMMMMMMMMMMMMMMMhmm.                                        -mdhMMMMMMMMMMMMm.
          yMMMMMMMMMMMMMMMMMN-                                          :NMMMMMMMMMMMMMMh
         /MMMMMMMMMMMMMMMMMN:                                            +MMMMMMMMMMMMMMM+
        'mMMMMMMMMMMMMMMMMMo                                              sMMMMMMMMMMMMMMN.
        oMMMMMMMMMMMMMMMMMh                                               'dMMMMMMMMMMMMMMy
       'mMMMMMMMMMMMMMMMMN.                                                :MMMMMMMMMMMMMMM-
       :MMMMMMMMMMMMMMMMMo                                                  yMMMMMMMMMMMMMMy
       sMMMMMMMMMMMMMMMMN'                                                  -MMMMMMMMMMMMMMN'
       dMMMMMMMMMMMMMMMMh                                                    mMMMMMMMMMMMMMM-
       mMMMMMMMMMMMMMMMMo                                                    yMMMMMMMMMMMMMM-
       mMMMMMMMMMMMMMMMMo                                                    sMMMMMMMMMMMMMM.
       hMMMMMMMMMMMMMMMMs                                                    hMMMMMMMMMMMMMN
       oMMMMMMMMMMMMMMMMy                                                    dMMMMMMMMMMMMMy
       .MMMMMMMMMMMMMMMMd                                                    NMMMMMMMMMMMMM-
        yMMMMMMMMMMMMMMMM'                                                  .MMMMMMMMMMMMMh
        .NMMMMMMMMMMMMMMM/                                                  oMMMMMMMMMMMMN-
         :NMMMMMMMMMMMMMMh                                                  mMMMMMMMMMMMMo
          /NMMMMMMMMMMMMMM-                                                /MMMMMMMMMMMMd'
           :NMMMMMMMMMMMMMh                                               'mMMMMMMMMMMMm.
            .hMMMMMMMMMMMMM/                                              oMMMMMMMMMMMN-
              +mMMMMMMMMMMMN-                                            :MMMMMMMMMMMm-
               'oNMMMMMMMMMMm.                                          -NMMMMMMMMMMd-
                 .omNmh+:hNMMm-                                        :NNsmMMMMMMMy'
                   '.     -smMN+                                     'oNh- 'sNMMNh:
                            ':yNh-                                  -hh:     .:-'
                               ':o/'                              '/+.
                                   '                              '

`

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}
Author:{{if .Author}} {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

// RunCommand runs cmd on file
func RunCommand(ctx context.Context, cmd string, args ...string) (string, error) {

	var c *exec.Cmd

	if ctx != nil {
		c = exec.CommandContext(ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}

	output, err := c.Output()
	if err != nil {
		return string(output), err
	}

	// check for exec context timeout
	if ctx != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out", cmd)
		}
	}

	return string(output), nil
}

func restartDock() error {
	ctx := context.Background()

	utils.Indent(log.Info)("restarting Dock")
	if _, err := RunCommand(ctx, "killall", "Dock"); err != nil {
		return errors.Wrap(err, "killing Dock process failed")
	}

	// let system settle
	time.Sleep(5 * time.Second)

	return nil
}

func removeOldDatabaseFiles(dbpath string) error {

	paths := []string{
		filepath.Join(dbpath, "db"),
		filepath.Join(dbpath, "db-shm"),
		filepath.Join(dbpath, "db-wal"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			utils.DoubleIndent(log.WithField("path", path).Warn)("file not found")
			continue
		}
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "removing file failed")
		}
		utils.DoubleIndent(log.WithField("path", path).Info)("removed old file")

	}

	return restartDock()
}

func savePath(confPath string, icloud bool) string {

	if icloud {

		iCloudPath, err := getiCloudDrivePath()
		if err != nil {
			log.WithError(err).Fatal("get iCloud drive path failed")
		}

		if len(confPath) > 0 {
			return filepath.Join(iCloudPath, confPath)
		}

		host, err := os.Hostname()
		if err != nil {
			log.WithError(err).Fatal("get hostname failed")
		}
		host = strings.TrimRight(host, ".local")

		return filepath.Join(iCloudPath, ".launchpad."+host+".yaml")
	}

	if len(confPath) > 0 {
		return confPath
	}

	// get current user
	user, err := user.Current()
	if err != nil {
		log.WithError(err).Fatal("get current user failed")
	}

	return filepath.Join(user.HomeDir, ".launchpad.yaml")
}

func getiCloudDrivePath() (string, error) {

	// get current user
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(user.HomeDir, "Library/Mobile Documents/com~apple~CloudDocs"), nil
}

func split(buf []string, lim int) [][]string {
	var chunk []string
	chunks := make([][]string, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}
