package models

type CoreSearchPrivateCourseResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SectionName string `json:"section_name"`
	Path        string `json:"path"`
}

type CoreSearchPublicCourseResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type CoreSearchSectionResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CourseName string `json:"course_name"`
	Path       string `json:"path"`
}

type CoreSearchLabResult struct {
	ID          string `json:"id"`
	LabName     string `json:"lab_name"`
	SectionName string `json:"section_name"`
	CourseName  string `json:"course_name"`
	Path        string `json:"path"`
}

type CoreSearchMaterialResult struct {
	ID           string `json:"id"`
	MaterialName string `json:"material_name"`
	LabName      string `json:"lab_name"`
	SectionName  string `json:"section_name"`
	CourseName   string `json:"course_name"`
	Path         string `json:"path"`
}

type CoreSearchResult struct {
	PrivateCourses []CoreSearchPrivateCourseResult `json:"private_courses"`
	PublicCourses  []CoreSearchPublicCourseResult  `json:"public_courses"`
	Sections       []CoreSearchSectionResult       `json:"sections"`
	Labs           []CoreSearchLabResult           `json:"labs"`
	Materials      []CoreSearchMaterialResult      `json:"materials"`
}
