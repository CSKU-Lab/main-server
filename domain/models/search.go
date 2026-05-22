package models

type SearchCourseResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type SearchLabResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CourseID   string `json:"course_id"`
	CourseName string `json:"course_name"`
	Path       string `json:"path"`
}

type SearchMaterialResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	CourseID   string `json:"course_id"`
	CourseName string `json:"course_name"`
	Path       string `json:"path"`
}

type SearchSectionResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CourseID   string `json:"course_id"`
	CourseName string `json:"course_name"`
	Path       string `json:"path"`
}

type SearchSectionLabResult struct {
	ID          string `json:"id"`
	LabName     string `json:"lab_name"`
	SectionName string `json:"section_name"`
	CourseName  string `json:"course_name"`
	Path        string `json:"path"`
}

type SearchSectionLabMaterialResult struct {
	ID           string `json:"id"`
	MaterialName string `json:"material_name"`
	LabName      string `json:"lab_name"`
	SectionName  string `json:"section_name"`
	CourseName   string `json:"course_name"`
	Path         string `json:"path"`
}

type SearchResult struct {
	Courses             []SearchCourseResult             `json:"courses"`
	Labs                []SearchLabResult                `json:"labs"`
	Materials           []SearchMaterialResult           `json:"materials"`
	Sections            []SearchSectionResult            `json:"sections"`
	SectionLabs         []SearchSectionLabResult         `json:"section_labs"`
	SectionLabMaterials []SearchSectionLabMaterialResult `json:"section_lab_materials"`
}
