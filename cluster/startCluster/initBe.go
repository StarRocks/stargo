
package startCluster

import(
    "fmt"
    "time"
    "errors"
    "stargo/sr-utl"
    "stargo/module"
    "stargo/cluster/checkStatus"
)




func InitBeCluster(yamlConf *module.ConfStruct) {

    var infoMess string
    var err error
    var beStat map[string]string

    // start Fe node one by one
    var tmpUser string
    var tmpKeyRsa string
    var tmpSshHost string
    var tmpSshPort int
    var tmpHeartbeatServicePort int
    var tmpBeDeployDir string
    var beStatusList string
    // var tmpFeEntryHost string
    // var tmpFeEntryPort int
    tmpUser = module.GYamlConf.Global.User
    tmpKeyRsa = module.GSshKeyRsa

    // get FE entry
    feEntryId, err := checkStatus.GetFeEntry(-1)
    //tmpFeEntryHost = yamlConf.FeServers[feEntryId].Host
    //tmpFeEntryPort = yamlConf.FeServers[feEntryId].QueryPort
    module.SetFeEntry(feEntryId)
    if err != nil || feEntryId == -1 {
        infoMess = "Error in get the FE entry, pls check FE status."
	utl.Log("ERROR", infoMess)
	err = errors.New(infoMess)
	panic(err)
    }



    for i := 0; i < len(yamlConf.BeServers); i++ {

        tmpSshHost = yamlConf.BeServers[i].Host
        tmpSshPort = yamlConf.BeServers[i].SshPort
        tmpHeartbeatServicePort = yamlConf.BeServers[i].HeartbeatServicePort
        tmpBeDeployDir = yamlConf.BeServers[i].DeployDir

	infoMess = fmt.Sprintf("Starting BE node [BeHost = %s HeartbeatServicePort = %d]", tmpSshHost, tmpHeartbeatServicePort)
        utl.Log("INFO", infoMess)

	for startTimeInd := 0; startTimeInd < 3; startTimeInd++ {

	    infoMess = fmt.Sprintf("The %d time to start [%s]",(startTimeInd + 1), tmpSshHost)
            utl.Log("DEBUG", infoMess)
	    // startBeNode(user string, keyRsa string, sshHost string, sshPort int, heartbeatServicePort int, beDeployDir string) (err error)
	    err = initBeNode(tmpUser, tmpKeyRsa, tmpSshHost, tmpSshPort, tmpHeartbeatServicePort, tmpBeDeployDir)

	    startWaitTime := time.Duration(20 - startTimeInd * 5)
	    // the be process need 20s to startup
	    time.Sleep(startWaitTime  * time.Second)

            beStat, _ = checkStatus.CheckBeStatus(i)
            if beStat["Alive"] == "true" {
                infoMess = fmt.Sprintf("The BE node start succefully [host = %s, heartbeatServicePort = %d]", tmpSshHost, tmpHeartbeatServicePort)
                utl.Log("INFO", infoMess)
                break
            } else {
                infoMess = fmt.Sprintf("The BE node doesn't start, wait for 10s [BeHost = %s, HeartbeatServicePort = %d, error = %v]", tmpSshHost, tmpHeartbeatServicePort, err)
                utl.Log("WARN", infoMess)
            }
        } // FOR-END: 3 time to restart BE node

	if beStat["Alive"] == "false" {
             infoMess = fmt.Sprintf("The BE node start failed [BeHost = %s, HeartbeatServicePort = %d, error = %v]", tmpSshHost, tmpHeartbeatServicePort, err)
        }

	beStatusList = beStatusList + "                                        " + fmt.Sprintf("beHost = %-20sbeHeartbeatServicePort = %d\tbeStatus = %v\n", tmpSshHost, tmpHeartbeatServicePort, beStat["Alive"])
    }
    beStatusList = "List all BE status:\n" + beStatusList
    utl.Log("OUTPUT", beStatusList)
}

func initBeNode(user string, keyRsa string, sshHost string, sshPort int, heartbeatServicePort int, beDeployDir string) (err error) {

    var infoMess string


    addBeSQL := fmt.Sprintf("alter system add backend \"%s:%d\"", sshHost, heartbeatServicePort)
    addBeCMD := fmt.Sprintf("%s/bin/start_be.sh --daemon", beDeployDir)

    //infoMess = fmt.Sprintf("Starting BE node [host = %s, heartbeatServicePort = %d]", sshHost, heartbeatServicePort)
    //utl.Log("INFO", infoMess)

    // alter system add backend "sshHost:heartbeatServicePort"
    sqlUserName := "root"
    sqlPassword := ""
    sqlIp := module.GFeEntryHost
    sqlPort := module.GFeEntryQueryPort
    sqlDbName := ""

    _, err = utl.RunSQL(sqlUserName, sqlPassword, sqlIp, sqlPort, sqlDbName, addBeSQL)
    if err != nil {
        infoMess = fmt.Sprintf(`Error in add follower BE node, [
                                        sqlUserName = %s
                                        sqlPassword = %s
                                        sqlIP = %s
                                        sqlPort = %d
                                        sqlDBName = %s
                                        addFollowerSQL =%s
                                        errMess = %v]`, sqlUserName, sqlPassword, sqlIp, sqlPort, sqlDbName, addBeSQL, err)
        utl.Log("ERROR", infoMess)
        return err
    }

    // run beDeploy/bin/start_be.sh --daemon 
    _, err = utl.SshRun(user, keyRsa, sshHost, sshPort, addBeCMD)
    if err != nil {
        infoMess = fmt.Sprintf(`Waiting for startMastertFeNode:
                                        user = %s
                                        keyRsa = %s
                                        sshHost = %s
                                        sshPort = %d
                                        beDeployDir = %s`,
                user, keyRsa, sshHost, sshPort, beDeployDir)
        utl.Log("WARN", infoMess)
        return err
    }

    // time.Sleep(5 * time.Second)
    return nil

}
