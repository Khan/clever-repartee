package rostering

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Khan/clever-repartee/pkg/tripperware"

	"go.uber.org/zap"

	"github.com/Khan/clever-repartee/pkg/generated"
	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
)

func GetCleverClient(
	logger *zap.Logger,
	districtID string,
	isMAP bool,
) (*generated.Client, error) {
	districtToken, err := GetCleverToken(logger, districtID, isMAP)
	if err != nil {
		return nil, err
	}
	bearerTokenProvider, bearerTokenProviderErr := securityprovider.
		NewSecurityProviderBearerToken(districtToken)
	if bearerTokenProviderErr != nil {
		panic(bearerTokenProviderErr)
	}

	pesterClient := tripperware.NewLoggedRetryHTTPClient(logger)

	client, clientErr := generated.NewClient(
		"https://api.clever.com/v2.1/", []generated.ClientOption{
			generated.WithRequestEditorFn(bearerTokenProvider.Intercept),
			generated.WithHTTPClient(pesterClient),
		}...,
	)
	return client, clientErr
}

func ParseLinkStartingAfter(nextLink string) string {
	u, err := url.Parse(nextLink)
	if err != nil {
		return ""
	}
	m, _ := url.ParseQuery(u.RawQuery)
	startingAfter := m["starting_after"]
	if len(startingAfter) > 0 {
		return startingAfter[0]
	}
	return ""
}

func GetCleverDistricts(
	client *generated.Client,
) (*[]generated.District, error) {
	var districts []generated.District
	districtsParams := &generated.GetDistrictsParams{}

	resp, err := client.GetDistricts(
		context.Background(), //nolint:ka-context // GKE ≠ AppEngine
		districtsParams,
	)
	if err != nil {
		return nil, err
	}

	if !IsHTTPSuccess(resp.StatusCode) {
		resp.Body.Close()
		return nil, fmt.Errorf(
			"HTTP %d Error for Clever Request /districts",
			resp.StatusCode,
		)
	}

	districtsResp := generated.DistrictsResponse{}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	err = dec.Decode(&districtsResp)
	if err != nil {
		return nil, err
	}

	if districtsResp.Data == nil {
		return &districts, nil
	}

	data := *districtsResp.Data
	for i := range data {
		districts = append(districts, *data[i].Data)
	}
	return &districts, nil
}

func GetCleverSchools(
	client *generated.Client,
	limit int,
) (*[]generated.School, error) {
	var schools []generated.School
	schoolsParams := &generated.GetSchoolsParams{Limit: &limit}
	next := true
	for next {
		resp, err := client.GetSchools(
			context.Background(), //nolint:ka-context // GKE ≠ AppEngine
			schoolsParams,
		)
		if err != nil {
			return nil, err
		}
		next = false
		if !IsHTTPSuccess(resp.StatusCode) {
			resp.Body.Close()
			return &schools, fmt.Errorf(
				"HTTP %d Error for Clever Request /schools starting after %s",
				resp.StatusCode, *schoolsParams.StartingAfter,
			)
		}

		schoolsResp := &generated.SchoolsResponse{}

		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(schoolsResp)
		if err != nil {
			return nil, err
		}
		if schoolsResp.Data != nil {
			data := *schoolsResp.Data

			for i := range data {
				schools = append(schools, *data[i].Data)
			}
		}
		if schoolsResp.Links != nil {
			links := *schoolsResp.Links

			for i := range links {
				if *links[i].Rel == "next" {
					next = true
					sa := ParseLinkStartingAfter(*links[i].Uri)
					schoolsParams.StartingAfter = &sa
				}
			}
		}
	}
	return &schools, nil
}

func GetCleverDistrictAdmins(
	client *generated.Client,
	limit int,
) (*[]generated.DistrictAdmin, error) {
	var districtAdmins []generated.DistrictAdmin
	districtAdminsParams := &generated.GetDistrictAdminsParams{Limit: &limit}
	next := true
	for next {
		resp, err := client.GetDistrictAdmins(
			context.Background(), //nolint:ka-context // GKE ≠ AppEngine
			districtAdminsParams,
		)
		if err != nil {
			return nil, err
		}

		next = false
		if !IsHTTPSuccess(resp.StatusCode) {
			resp.Body.Close()
			return &districtAdmins, fmt.Errorf(
				"HTTP %d Error for Clever Request /district_admins starting after %s",
				resp.StatusCode,
				*districtAdminsParams.StartingAfter,
			)
		}
		districtAdminsResp := &generated.DistrictAdminsResponse{}

		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(districtAdminsResp)
		if err != nil {
			return nil, err
		}

		data := *districtAdminsResp.Data

		for i := range data {
			districtAdmins = append(districtAdmins, *data[i].Data)
		}
		links := *districtAdminsResp.Links
		for i := range links {
			if *links[i].Rel == "next" {
				next = true
				sa := ParseLinkStartingAfter(*links[i].Uri)
				districtAdminsParams.StartingAfter = &sa
			}
		}
	}

	return &districtAdmins, nil
}

func GetCleverStudents(
	client *generated.Client,
	limit int,
) (*[]generated.Student, error) {
	var students []generated.Student
	studentsParams := &generated.GetStudentsParams{Limit: &limit}
	next := true
	for next {
		resp, err := client.GetStudents(
			context.Background(), //nolint:ka-context // GKE ≠ AppEngine
			studentsParams,
		)
		if err != nil {
			return nil, err
		}

		next = false
		if !IsHTTPSuccess(resp.StatusCode) {
			resp.Body.Close()
			return &students, fmt.Errorf(
				"HTTP %d Error for Clever Request /students starting after %s",
				resp.StatusCode, *studentsParams.StartingAfter,
			)
		}
		studentsResp := &generated.StudentsResponse{}

		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(studentsResp)
		if err != nil {
			return nil, err
		}
		if studentsResp.Data != nil {
			data := *studentsResp.Data

			for i := range data {
				students = append(students, *data[i].Data)
			}
		}
		if studentsResp.Links != nil {
			links := *studentsResp.Links

			for i := range links {
				if *links[i].Rel == "next" {
					next = true
					sa := ParseLinkStartingAfter(*links[i].Uri)
					studentsParams.StartingAfter = &sa
				}
			}
		}
	}

	return &students, nil
}

func GetCleverTeachers(
	client *generated.Client,
	limit int,
) (*[]generated.Teacher, error) {
	var teachers []generated.Teacher
	teachersParams := &generated.GetTeachersParams{Limit: &limit}
	next := true
	for next {
		resp, err := client.GetTeachers(
			context.Background(), //nolint:ka-context // GKE ≠ AppEngine
			teachersParams,
		)
		if err != nil {
			return nil, err
		}

		next = false
		if !IsHTTPSuccess(resp.StatusCode) {
			resp.Body.Close()
			return &teachers, fmt.Errorf(
				"HTTP %d Error for Clever Request /teachers starting after %s",
				resp.StatusCode, *teachersParams.StartingAfter,
			)
		}
		teachersResp := &generated.TeachersResponse{}
		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(teachersResp)
		if err != nil {
			return nil, err
		}
		data := *teachersResp.Data

		if teachersResp.Data != nil {
			for i := range data {
				teachers = append(teachers, *data[i].Data)
			}
		}
		if teachersResp.Links != nil {
			links := *teachersResp.Links

			for i := range links {
				if *links[i].Rel == "next" {
					next = true
					sa := ParseLinkStartingAfter(*links[i].Uri)
					teachersParams.StartingAfter = &sa
				}
			}
		}
	}

	return &teachers, nil
}

func GetCleverSchoolAdmins(
	client *generated.Client,
	limit int,
) (*[]generated.SchoolAdmin, error) {
	var schoolAdmins []generated.SchoolAdmin
	schoolAdminsParams := &generated.GetSchoolAdminsParams{Limit: &limit}
	next := true
	for next {
		resp, err := client.GetSchoolAdmins(
			context.Background(), //nolint:ka-context // GKE ≠ AppEngine
			schoolAdminsParams,
		)
		if err != nil {
			return nil, err
		}

		next = false
		if !IsHTTPSuccess(resp.StatusCode) {
			resp.Body.Close()
			return &schoolAdmins, fmt.Errorf(
				"HTTP %d Error for Clever Request /school_admins starting after %s",
				resp.StatusCode,
				*schoolAdminsParams.StartingAfter,
			)
		}

		schoolAdminsResp := &generated.SchoolAdminsResponse{}
		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(schoolAdminsResp)
		if err != nil {
			return nil, err
		}

		if schoolAdminsResp.Data != nil {
			data := *schoolAdminsResp.Data
			for i := range data {
				schoolAdmins = append(schoolAdmins, *data[i].Data)
			}
		}
		if schoolAdminsResp.Links != nil {
			links := *schoolAdminsResp.Links
			for i := range links {
				if *links[i].Rel == "next" {
					next = true
					sa := ParseLinkStartingAfter(*links[i].Uri)
					schoolAdminsParams.StartingAfter = &sa
				}
			}
		}
	}

	return &schoolAdmins, nil
}

func GetCleverSections(
	client *generated.Client,
	limit int,
) (*[]generated.Section, error) {
	var sections []generated.Section
	sectionsParams := &generated.GetSectionsParams{Limit: &limit}
	next := true
	for next {
		resp, err := client.GetSections(
			context.Background(), //nolint:ka-context // GKE ≠ AppEngine
			sectionsParams,
		)
		if err != nil {
			return nil, err
		}

		next = false
		if !IsHTTPSuccess(resp.StatusCode) {
			resp.Body.Close()
			return &sections, fmt.Errorf(
				"HTTP %d Error for Clever Request /sections starting after %s",
				resp.StatusCode, *sectionsParams.StartingAfter,
			)
		}

		sectionsResp := &generated.SectionsResponse{}
		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(sectionsResp)
		if err != nil {
			return nil, err
		}
		if sectionsResp.Data != nil {
			data := *sectionsResp.Data
			for i := range data {
				sections = append(sections, *data[i].Data)
			}
		}
		if sectionsResp.Links != nil {
			links := *sectionsResp.Links
			for i := range links {
				if *links[i].Rel == "next" {
					next = true
					sa := ParseLinkStartingAfter(*links[i].Uri)
					sectionsParams.StartingAfter = &sa
				}
			}
		}
	}

	return &sections, nil
}

func IsHTTPSuccess(code int) bool {
	return code >= 200 && code <= 299
}
