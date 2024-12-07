package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

func CreateSnapshot(cfg aws.Config, dbInstanceID, snapshotName string) {

	log.Println("\033[1;32m[+]\033[0m RDS: Creating Snapshot:", snapshotName, "| InstanceID:", dbInstanceID)
	client := rds.NewFromConfig(cfg)

	input := &rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: &dbInstanceID,
		DBSnapshotIdentifier: &snapshotName,
	}

	_, err := client.CreateDBSnapshot(context.TODO(), input)
	if err != nil {
		log.Println(err)
		log.Println("\033[1;31m[-]\033[0m RDS: Snapshot creation failed, skipping...")
	}
}

func CopyDBSnapshot(cfg aws.Config, sourceSnapshotID string, targetSnapshotID string, kmsKeyId string, region string) {

	client := rds.NewFromConfig(cfg)
	var err error

	for {
		log.Println("\033[1;32m[+]\033[0m RDS: Copying Snapshot:", sourceSnapshotID, "| KMS:", kmsKeyId)
		_, err = client.CopyDBSnapshot(context.TODO(), &rds.CopyDBSnapshotInput{
			SourceDBSnapshotIdentifier: &sourceSnapshotID,
			TargetDBSnapshotIdentifier: &targetSnapshotID,
			KmsKeyId:                   &kmsKeyId,
			SourceRegion:               &region,
		})

		if err != nil {
			fmt.Println(err)
			if strings.Contains(err.Error(), "already exists") {
				log.Println("\033[1;31m[-]\033[0m RDS: Snapshot already exists, skipping...")
				break
			}
		}

		break
	}

}

func ModifyDBSnapshotAttribute(cfg aws.Config, snapshotID string, accountID string) {

	log.Println("\033[1;32m[+]\033[0m RDS: Modifying Snapshot:", snapshotID, "| AccountID:", accountID)
	client := rds.NewFromConfig(cfg)

	params := &rds.ModifyDBSnapshotAttributeInput{
		DBSnapshotIdentifier: &snapshotID,
		AttributeName:        aws.String("restore"),
		ValuesToAdd:          []string{accountID},
	}

	_, err := client.ModifyDBSnapshotAttribute(context.TODO(), params)
	if err != nil {
		log.Fatalf("unable to share snapshot, %v", err)
	}

	log.Println("\033[1;32m[+]\033[0m RDS: Snapshot:", snapshotID, "has been shared with account:", accountID)
}

func DescribeDBSnapshots(session aws.Config, snapshotID string, action string) {

	client := rds.NewFromConfig(session)

	params := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: &snapshotID,
	}

	for {
		result, err := client.DescribeDBSnapshots(context.TODO(), params)
		status := result.DBSnapshots[0].Status

		if err != nil {
			log.Fatalln(err)
		}

		if *status != "available" {

			if action == "copy" {
				log.Println("\033[1;32m[+]\033[0m RDS: Waiting for snapshot to be copied...")
			} else {
				log.Println("\033[1;32m[+]\033[0m RDS: Waiting for snapshot to be created...")
			}

			time.Sleep(time.Second * 60)
		} else {
			break
		}
	}

}

func DescribeDBInstances(session aws.Config, snapshotID string) {

	client := rds.NewFromConfig(session)

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: &snapshotID,
	}

	for {

		result, err := client.DescribeDBInstances(context.TODO(), params)
		dbname := result.DBInstances[0].DBInstanceIdentifier
		status := result.DBInstances[0].DBInstanceStatus

		if err != nil {
			log.Fatalln(err)
		}

		if *status != "available" {

			log.Println("\033[1;32m[+]\033[0m RDS: Waiting for DBInstance to be restored...")

			time.Sleep(time.Second * 60)
		} else {
			log.Println("\033[1;32m[+]\033[0m RDS: DBInstance:", *dbname, "has been restored!")
			break
		}
	}

}

func RestoreDBInstanceFromDBSnapshot(session aws.Config, snapshotID string, dbInstanceName string, region string) {

	log.Println("\033[1;32m[+]\033[0m RDS: Restoring DB Instance from Snapshot:", snapshotID, "| Region:", region)
	client := rds.NewFromConfig(session)

	_, err := client.RestoreDBInstanceFromDBSnapshot(context.TODO(), &rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceIdentifier: &dbInstanceName,
		DBSnapshotIdentifier: &snapshotID,
	})

	if err != nil {
		log.Println(err)
		log.Println("\033[1;31m[-]\033[0m RDS: DBInstance:", dbInstanceName, "already exist, skipping!")
	}

}
