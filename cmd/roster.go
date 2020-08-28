package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"

	"github.com/Khan/clever-repartee/pkg/generated"
	"github.com/Khan/clever-repartee/pkg/mail"
	"github.com/Khan/clever-repartee/pkg/rostering"
)

func DiffCommand(logger *zap.Logger) *Command {
	cmd := &Command{
		UsageLine: "diff",
		Short:     "Compare a district's roster via two different Clever apps",
		Long:      "Compare a district's roster with Clever ID for -district flag via two different Clever apps",
		Run:       Diff,
		Logger:    logger,
	}
	return cmd
}

func Diff(cmd *Command, _ []string) error {
	var districtCleverID string

	var writeJson bool

	flag.StringVar(&districtCleverID, "district", "", "District Clever ID")
	flag.BoolVar(&writeJson, "json", false, "Write JSON files to local disk")

	flagErr := flag.CommandLine.Parse(getFlags())
	if flagErr != nil {
		return flagErr
	}

	if districtCleverID == "" {
		return fmt.Errorf("-district ${ID} is a required argument")
	}

	logger := cmd.Logger

	logger.Info(
		fmt.Sprintf(
			"Processing district with clever ID %s!\n",
			districtCleverID,
		))

	var districtName string
	mapAcceleratorCleverClient, mapAcceleratorClientErr := rostering.GetCleverClient(
		logger,
		districtCleverID,
		false,
	)
	if mapAcceleratorClientErr != nil {
		return mapAcceleratorClientErr
	}

	mapAcceleratorRoster, mapAcceleratorRosterErr := GetRoster(
		logger,
		mapAcceleratorCleverClient,
	)
	if mapAcceleratorRosterErr != nil {
		return mapAcceleratorRosterErr
	}

	mapGrowthCleverClient, mapGrowthClientErr := rostering.GetCleverClient(
		logger,
		districtCleverID,
		true,
	)
	if mapGrowthClientErr != nil {
		return mapGrowthClientErr
	}

	mapGrowthRoster, mapGrowthRosterErr := GetRoster(
		logger,
		mapGrowthCleverClient,
	)
	if mapGrowthRosterErr != nil {
		return mapGrowthRosterErr
	}
	for i := range *mapAcceleratorRoster.districts {
		if mapAcceleratorRoster.districts != nil {
			if (*mapAcceleratorRoster.districts)[i].Name != nil {
				districtName = *(*mapAcceleratorRoster.districts)[i].Name
			}
		}
	}

	missingStudents := findMissingStudents(
		mapGrowthRoster,
		mapAcceleratorRoster,
	)
	missingTeachers := findMissingTeachers(
		mapGrowthRoster,
		mapAcceleratorRoster,
	)
	missingSchools := findMissingSchools(mapGrowthRoster, mapAcceleratorRoster)

	missingReport := mail.MissingReport{
		DistrictName:            districtName,
		DistrictCleverID:        districtCleverID,
		MissingStudentCleverIDs: missingStudents,
		MissingTeacherCleverIDs: missingTeachers,
		MissingSchoolCleverIDs:  missingSchools,
	}

	fromEmail := os.Getenv("FROM_EMAIL") //gmail account
	toEmail := os.Getenv("TO_EMAIL")
	password := os.Getenv("GMAIL_PASSWORD")
	host := "smtp.gmail.com"
	port := "587"
	subject := "=?utf-8?Q?=F0=9F=95=B5=EF=B8=8F?= Clever Discrepancy Report"

	bodyMessage, bodyErr := mail.NewSummaryMailBody(&missingReport)
	if bodyErr != nil {
		logger.Error(
			"Unable to compose summary email message body",
			zap.Error(bodyErr),
		)
	}
	mailErr := mail.Mail(
		fromEmail,
		toEmail,
		password,
		host,
		port,
		subject,
		bodyMessage,
	)
	if mailErr != nil {
		logger.Error(
			"Unable to send summary email message",
			zap.Error(bodyErr),
		)
	}

	// For local testing/debugging since transient files will be lost in
	// GKE job
	if writeJson {
		jsonWriteErr := writeDistrictToJSON(
			&missingReport,
		)
		if jsonWriteErr != nil {
			return jsonWriteErr
		}
	}

	return nil
}

// note not defensively nil safe, but in practice probably ok?
func findMissingStudents(
	mapGrowthRoster *Roster,
	mapAcceleratorRoster *Roster,
) []string {
	mapGrowthStudentIDmap := map[string]bool{}
	for i := range *mapGrowthRoster.students {
		mapGrowthStudent := (*mapGrowthRoster.students)[i]
		mapGrowthStudentIDmap[*mapGrowthStudent.Id] = true
	}
	var missingStudents []string
	for i := range *mapAcceleratorRoster.students {
		mapAcceleratorStudent := (*mapGrowthRoster.students)[i]
		if _, ok := mapGrowthStudentIDmap[*mapAcceleratorStudent.Id]; !ok {
			missingStudents = append(missingStudents, *mapAcceleratorStudent.Id)
		}
	}
	return missingStudents
}

// note not defensively nil safe, but in practice probably ok?
func findMissingTeachers(
	mapGrowthRoster *Roster,
	mapAcceleratorRoster *Roster,
) []string {
	mapGrowthTeacherIDmap := map[string]bool{}
	for i := range *mapGrowthRoster.teachers {
		mapGrowthTeacher := (*mapGrowthRoster.teachers)[i]
		mapGrowthTeacherIDmap[*mapGrowthTeacher.Id] = true
	}
	var missing []string
	for i := range *mapAcceleratorRoster.teachers {
		mapAcceleratorTeacher := (*mapGrowthRoster.teachers)[i]
		if _, ok := mapGrowthTeacherIDmap[*mapAcceleratorTeacher.Id]; !ok {
			missing = append(missing, *mapAcceleratorTeacher.Id)
		}
	}
	return missing
}

// note not defensively nil safe, but in practice probably ok?
func findMissingSchools(
	mapGrowthRoster *Roster,
	mapAcceleratorRoster *Roster,
) []string {
	mapGrowthSchoolIDmap := map[string]bool{}
	for i := range *mapGrowthRoster.schools {
		mapGrowthSchool := (*mapGrowthRoster.schools)[i]
		mapGrowthSchoolIDmap[*mapGrowthSchool.Id] = true
	}
	var missing []string
	for i := range *mapAcceleratorRoster.teachers {
		mapAcceleratorSchool := (*mapGrowthRoster.schools)[i]
		if _, ok := mapGrowthSchoolIDmap[*mapAcceleratorSchool.Id]; !ok {
			missing = append(missing, *mapAcceleratorSchool.Id)
		}
	}
	return missing
}

func writeDistrictToJSON(report *mail.MissingReport) error {
	file, err := json.MarshalIndent(*report, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile((*report).DistrictCleverID+".json", file, 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetRoster(
	logger *zap.Logger,
	clientClever *generated.Client,
) (*Roster, error) {
	roster := Roster{}

	districts, distErr := rostering.GetCleverDistricts(clientClever)
	if distErr != nil {
		return nil, distErr
	}
	roster.districts = districts

	schools, schoolErr := rostering.GetCleverSchools(clientClever, 1000)
	if schoolErr != nil {
		return nil, schoolErr
	}
	roster.schools = schools

	students, studentErr := rostering.GetCleverStudents(clientClever, 1000)
	if studentErr != nil {
		return nil, studentErr
	}
	roster.students = students

	teachers, teachErr := rostering.GetCleverTeachers(clientClever, 1000)
	if teachErr != nil {
		return nil, teachErr
	}
	roster.teachers = teachers

	districtAdmins, distAdmErr := rostering.GetCleverDistrictAdmins(
		clientClever,
		1000,
	)
	if distAdmErr != nil {
		return nil, distAdmErr
	}
	roster.districtAdmins = districtAdmins

	schoolAdmins, schoolAdminErr := rostering.GetCleverSchoolAdmins(
		clientClever,
		1000,
	)
	if schoolAdminErr != nil {
		return nil, schoolAdminErr
	}
	roster.schoolAdmins = schoolAdmins

	sections, sectionErr := rostering.GetCleverSections(clientClever, 1000)
	if sectionErr != nil {
		return nil, sectionErr
	}
	roster.sections = sections

	return &roster, nil
}

type Roster struct {
	districts      *[]generated.District
	schools        *[]generated.School
	students       *[]generated.Student
	teachers       *[]generated.Teacher
	districtAdmins *[]generated.DistrictAdmin
	schoolAdmins   *[]generated.SchoolAdmin
	sections       *[]generated.Section
}
