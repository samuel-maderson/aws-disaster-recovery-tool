package main

import (
	"aws-disaster-recovery-tool/src/resources/config"
	"aws-disaster-recovery-tool/src/resources/rds"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/akamensky/argparse"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/joho/godotenv"
)

var (
	dbID            string
	snapID          string
	snapIDEncrypted string
	session         aws.Config
	kmsKeyId        string
	region          string
	backupAccountID string
	accountID       string
	parser          *argparse.Parser
	e               *string
	envFile         string
)

func init() {

	t := time.Now()
	date := t.Format("01-02-2006")

	parser = argparse.NewParser("AWS Disaster Recovery Tool", "")
	e = parser.String("e", "env", &argparse.Options{Required: true, Help: "AWS Environment: production or backup"})
	err := parser.Parse(os.Args)

	if err != nil {
		log.Fatalln("\033[1;31m[-]\033[0m ", parser.Usage(err))
	}

	if *e == "production" {
		envFile = ".env.production"

	} else if *e == "backup" {
		envFile = ".env.backup"
	}

	err = godotenv.Load(envFile)

	if err != nil {
		log.Fatalln(err)
	}

	dbID = os.Getenv("APP_AWS_RDS_NAME")
	snapID = fmt.Sprintf("%s%s", os.Getenv("APP_AWS_SNAPSHOT_NAME"), date)
	snapIDEncrypted = fmt.Sprintf("%s%s", os.Getenv("APP_AWS_SNAPSHOT_ENCRYPTED_NAME"), date)
	kmsKeyId = os.Getenv("APP_AWS_KMS_ID")
	backupAccountID = os.Getenv("APP_AWS_BACKUP_ACCOUNT_ID")
	accountID = os.Getenv("APP_AWS_ACCOUNT_ID")
	region = os.Getenv("AWS_DEFAULT_REGION")

	session = config.AWS_SESSION(os.Getenv("AWS_DEFAULT_REGION"))

}

func main() {

	if *e == "production" {
		log.Println("\033[1;32m[+]\033[0m Starting process | [Env]:", *e, "| [Account ID]:", accountID, "| [Region]:", region)
		rds.CreateSnapshot(session, dbID, snapID)
		rds.DescribeDBSnapshots(session, snapID, "create")
		rds.CopyDBSnapshot(session, snapID, snapIDEncrypted, kmsKeyId, region)
		rds.DescribeDBSnapshots(session, snapIDEncrypted, "copy")
		rds.ModifyDBSnapshotAttribute(session, snapIDEncrypted, backupAccountID)
	} else if *e == "backup" {
		log.Println("\033[1;32m[+]\033[0m Starting process | [Env]:", *e, "| [Account ID]:", accountID, "| [Region]:", region)
		rds.CopyDBSnapshot(session, snapID, snapIDEncrypted, kmsKeyId, region)
		rds.DescribeDBSnapshots(session, snapIDEncrypted, "copy")
		rds.RestoreDBInstanceFromDBSnapshot(session, snapIDEncrypted, dbID, region)
		rds.DescribeDBInstances(session, dbID)
	}
}
