package sqlx

import (
	"context"
	"fmt"
	"sync"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type sqlxSearchRepository struct {
	db *sqlx.DB
}

func NewSearchRepository(db *sqlx.DB) repositories.SearchRepository {
	return &sqlxSearchRepository{db: db}
}

func (r *sqlxSearchRepository) Search(ctx context.Context, q string, limit int) (*models.SearchResult, error) {
	result := &models.SearchResult{
		Courses:             []models.SearchCourseResult{},
		Labs:                []models.SearchLabResult{},
		Materials:           []models.SearchMaterialResult{},
		Sections:            []models.SearchSectionResult{},
		SectionLabs:         []models.SearchSectionLabResult{},
		SectionLabMaterials: []models.SearchSectionLabMaterialResult{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make([]error, 6)

	wg.Add(6)

	// Courses
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT id, name
			FROM courses
			WHERE similarity(name, $1) > 0.1 AND is_deleted = false
			ORDER BY similarity(name, $1) DESC
			LIMIT $2
		`, q, limit)
		if err != nil {
			errs[0] = fmt.Errorf("courses: %w", err)
			return
		}
		defer rows.Close()
		var out []models.SearchCourseResult
		for rows.Next() {
			var row struct {
				ID   string `db:"id"`
				Name string `db:"name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[0] = fmt.Errorf("courses scan: %w", err)
				return
			}
			out = append(out, models.SearchCourseResult{
				ID:   row.ID,
				Name: row.Name,
				Path: "/cms/courses/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Courses = out
		}
		mu.Unlock()
	}()

	// Labs
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT l.id, l.display_name AS name, l.course_id, c.name AS course_name
			FROM labs l
			JOIN courses c ON c.id = l.course_id
			WHERE similarity(l.display_name, $1) > 0.1
			  AND l.is_deleted = false AND c.is_deleted = false
			ORDER BY similarity(l.display_name, $1) DESC
			LIMIT $2
		`, q, limit)
		if err != nil {
			errs[1] = fmt.Errorf("labs: %w", err)
			return
		}
		defer rows.Close()
		var out []models.SearchLabResult
		for rows.Next() {
			var row struct {
				ID         string `db:"id"`
				Name       string `db:"name"`
				CourseID   string `db:"course_id"`
				CourseName string `db:"course_name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[1] = fmt.Errorf("labs scan: %w", err)
				return
			}
			out = append(out, models.SearchLabResult{
				ID:         row.ID,
				Name:       row.Name,
				CourseID:   row.CourseID,
				CourseName: row.CourseName,
				Path:       "/cms/courses/" + row.CourseID + "/labs/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Labs = out
		}
		mu.Unlock()
	}()

	// Materials
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT m.id, m.name, m.type, m.course_id, c.name AS course_name
			FROM materials m
			JOIN courses c ON c.id = m.course_id
			WHERE similarity(m.name, $1) > 0.1
			  AND m.is_deleted = false AND c.is_deleted = false
			ORDER BY similarity(m.name, $1) DESC
			LIMIT $2
		`, q, limit)
		if err != nil {
			errs[2] = fmt.Errorf("materials: %w", err)
			return
		}
		defer rows.Close()
		var out []models.SearchMaterialResult
		for rows.Next() {
			var row struct {
				ID         string `db:"id"`
				Name       string `db:"name"`
				Type       string `db:"type"`
				CourseID   string `db:"course_id"`
				CourseName string `db:"course_name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[2] = fmt.Errorf("materials scan: %w", err)
				return
			}
			out = append(out, models.SearchMaterialResult{
				ID:         row.ID,
				Name:       row.Name,
				Type:       row.Type,
				CourseID:   row.CourseID,
				CourseName: row.CourseName,
				Path:       "/cms/courses/" + row.CourseID + "/materials/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Materials = out
		}
		mu.Unlock()
	}()

	// Sections
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT s.id, s.name, s.course_id, c.name AS course_name
			FROM sections s
			JOIN courses c ON c.id = s.course_id
			WHERE similarity(s.name, $1) > 0.1
			  AND s.is_deleted = false AND c.is_deleted = false
			ORDER BY similarity(s.name, $1) DESC
			LIMIT $2
		`, q, limit)
		if err != nil {
			errs[3] = fmt.Errorf("sections: %w", err)
			return
		}
		defer rows.Close()
		var out []models.SearchSectionResult
		for rows.Next() {
			var row struct {
				ID         string `db:"id"`
				Name       string `db:"name"`
				CourseID   string `db:"course_id"`
				CourseName string `db:"course_name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[3] = fmt.Errorf("sections scan: %w", err)
				return
			}
			out = append(out, models.SearchSectionResult{
				ID:         row.ID,
				Name:       row.Name,
				CourseID:   row.CourseID,
				CourseName: row.CourseName,
				Path:       "/cms/courses/" + row.CourseID + "/sections/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Sections = out
		}
		mu.Unlock()
	}()

	// Section Labs
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT
				l.id,
				l.display_name AS lab_name,
				s.name AS section_name,
				c.name AS course_name,
				s.course_id,
				s.id AS section_id
			FROM lab_sections ls
			JOIN labs l ON l.id = ls.lab_id
			JOIN sections s ON s.id = ls.section_id
			JOIN courses c ON c.id = s.course_id
			WHERE GREATEST(similarity(l.display_name, $1), similarity(s.name, $1)) > 0.1
			  AND l.is_deleted = false
			  AND ls.is_deleted = false
			  AND s.is_deleted = false
			  AND c.is_deleted = false
			ORDER BY GREATEST(similarity(l.display_name, $1), similarity(s.name, $1)) DESC
			LIMIT $2
		`, q, limit)
		if err != nil {
			errs[4] = fmt.Errorf("section_labs: %w", err)
			return
		}
		defer rows.Close()
		var out []models.SearchSectionLabResult
		for rows.Next() {
			var row struct {
				ID          string `db:"id"`
				LabName     string `db:"lab_name"`
				SectionName string `db:"section_name"`
				CourseName  string `db:"course_name"`
				CourseID    string `db:"course_id"`
				SectionID   string `db:"section_id"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[4] = fmt.Errorf("section_labs scan: %w", err)
				return
			}
			out = append(out, models.SearchSectionLabResult{
				ID:          row.ID,
				LabName:     row.LabName,
				SectionName: row.SectionName,
				CourseName:  row.CourseName,
				Path:        "/cms/courses/" + row.CourseID + "/sections/" + row.SectionID + "/labs/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.SectionLabs = out
		}
		mu.Unlock()
	}()

	// Section Lab Materials
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT
				m.id AS material_id,
				m.name AS material_name,
				l.id AS lab_id,
				l.display_name AS lab_name,
				s.id AS section_id,
				s.name AS section_name,
				c.id AS course_id,
				c.name AS course_name
			FROM lab_materials lm
			JOIN materials m ON m.id = lm.material_id
			JOIN labs l ON l.id = lm.lab_id
			JOIN lab_sections ls ON ls.lab_id = l.id
			JOIN sections s ON s.id = ls.section_id
			JOIN courses c ON c.id = s.course_id
			WHERE GREATEST(
				similarity(m.name, $1),
				similarity(l.display_name, $1),
				similarity(s.name, $1)
			) > 0.1
			  AND lm.is_deleted = false
			  AND m.is_deleted = false
			  AND l.is_deleted = false
			  AND ls.is_deleted = false
			  AND s.is_deleted = false
			  AND c.is_deleted = false
			ORDER BY GREATEST(
				similarity(m.name, $1),
				similarity(l.display_name, $1),
				similarity(s.name, $1)
			) DESC
			LIMIT $2
		`, q, limit)
		if err != nil {
			errs[5] = fmt.Errorf("section_lab_materials: %w", err)
			return
		}
		defer rows.Close()
		var out []models.SearchSectionLabMaterialResult
		for rows.Next() {
			var row struct {
				MaterialID   string `db:"material_id"`
				MaterialName string `db:"material_name"`
				LabID        string `db:"lab_id"`
				LabName      string `db:"lab_name"`
				SectionID    string `db:"section_id"`
				SectionName  string `db:"section_name"`
				CourseID     string `db:"course_id"`
				CourseName   string `db:"course_name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[5] = fmt.Errorf("section_lab_materials scan: %w", err)
				return
			}
			out = append(out, models.SearchSectionLabMaterialResult{
				ID:           row.MaterialID,
				MaterialName: row.MaterialName,
				LabName:      row.LabName,
				SectionName:  row.SectionName,
				CourseName:   row.CourseName,
				Path:         "/cms/courses/" + row.CourseID + "/sections/" + row.SectionID + "/labs/" + row.LabID + "/materials/" + row.MaterialID + "/submissions",
			})
		}
		mu.Lock()
		if out != nil {
			result.SectionLabMaterials = out
		}
		mu.Unlock()
	}()

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *sqlxSearchRepository) SearchForStudent(ctx context.Context, userID, q string, limit int) (*models.CoreSearchResult, error) {
	result := &models.CoreSearchResult{
		PrivateCourses: []models.CoreSearchPrivateCourseResult{},
		PublicCourses:  []models.CoreSearchPublicCourseResult{},
		Sections:       []models.CoreSearchSectionResult{},
		Labs:           []models.CoreSearchLabResult{},
		Materials:      []models.CoreSearchMaterialResult{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make([]error, 5)

	wg.Add(5)

	// Private courses: enrolled via sections → link to student's section
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT c.id, c.name, s.name AS section_name, s.id AS section_id, similarity(c.name, $2) AS sim
			FROM courses c
			JOIN sections s ON s.course_id = c.id
			JOIN section_students ss ON ss.section_id = s.id
			WHERE ss.student_id = $1
			  AND similarity(c.name, $2) > 0.1
			  AND ss.is_deleted = false
			  AND s.is_deleted = false
			  AND c.is_deleted = false
			ORDER BY sim DESC
			LIMIT $3
		`, userID, q, limit)
		if err != nil {
			errs[0] = fmt.Errorf("core private courses: %w", err)
			return
		}
		defer rows.Close()
		var out []models.CoreSearchPrivateCourseResult
		for rows.Next() {
			var row struct {
				ID          string  `db:"id"`
				Name        string  `db:"name"`
				SectionName string  `db:"section_name"`
				SectionID   string  `db:"section_id"`
				Sim         float64 `db:"sim"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[0] = fmt.Errorf("core private courses scan: %w", err)
				return
			}
			out = append(out, models.CoreSearchPrivateCourseResult{
				ID:          row.ID,
				Name:        row.Name,
				SectionName: row.SectionName,
				Path:        "/sections/" + row.SectionID,
			})
		}
		mu.Lock()
		if out != nil {
			result.PrivateCourses = out
		}
		mu.Unlock()
	}()

	// Public courses: enrolled via course_enrollments → link to course page
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT c.id, c.name
			FROM courses c
			JOIN course_enrollments ce ON ce.course_id = c.id
			WHERE ce.student_id = $1
			  AND similarity(c.name, $2) > 0.1
			  AND c.is_deleted = false
			ORDER BY similarity(c.name, $2) DESC
			LIMIT $3
		`, userID, q, limit)
		if err != nil {
			errs[1] = fmt.Errorf("core public courses: %w", err)
			return
		}
		defer rows.Close()
		var out []models.CoreSearchPublicCourseResult
		for rows.Next() {
			var row struct {
				ID   string `db:"id"`
				Name string `db:"name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[1] = fmt.Errorf("core public courses scan: %w", err)
				return
			}
			out = append(out, models.CoreSearchPublicCourseResult{
				ID:   row.ID,
				Name: row.Name,
				Path: "/courses/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.PublicCourses = out
		}
		mu.Unlock()
	}()

	// Enrolled sections
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT s.id, s.name, c.name AS course_name
			FROM sections s
			JOIN section_students ss ON ss.section_id = s.id
			JOIN courses c ON c.id = s.course_id
			WHERE ss.student_id = $1
			  AND similarity(s.name, $2) > 0.1
			  AND ss.is_deleted = false
			  AND s.is_deleted = false
			  AND c.is_deleted = false
			ORDER BY similarity(s.name, $2) DESC
			LIMIT $3
		`, userID, q, limit)
		if err != nil {
			errs[2] = fmt.Errorf("core sections: %w", err)
			return
		}
		defer rows.Close()
		var out []models.CoreSearchSectionResult
		for rows.Next() {
			var row struct {
				ID         string `db:"id"`
				Name       string `db:"name"`
				CourseName string `db:"course_name"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[2] = fmt.Errorf("core sections scan: %w", err)
				return
			}
			out = append(out, models.CoreSearchSectionResult{
				ID:         row.ID,
				Name:       row.Name,
				CourseName: row.CourseName,
				Path:       "/sections/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Sections = out
		}
		mu.Unlock()
	}()

	// Enrolled section labs
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT
				ls.lab_id AS id,
				l.display_name AS lab_name,
				s.name AS section_name,
				c.name AS course_name,
				s.id AS section_id
			FROM section_students ss
			JOIN sections s ON s.id = ss.section_id
			JOIN courses c ON c.id = s.course_id
			JOIN lab_sections ls ON ls.section_id = s.id
			JOIN labs l ON l.id = ls.lab_id
			WHERE ss.student_id = $1
			  AND similarity(l.display_name, $2) > 0.1
			  AND ss.is_deleted = false
			  AND s.is_deleted = false
			  AND c.is_deleted = false
			  AND ls.is_deleted = false
			  AND l.is_deleted = false
			ORDER BY similarity(l.display_name, $2) DESC
			LIMIT $3
		`, userID, q, limit)
		if err != nil {
			errs[3] = fmt.Errorf("core labs: %w", err)
			return
		}
		defer rows.Close()
		var out []models.CoreSearchLabResult
		for rows.Next() {
			var row struct {
				ID          string `db:"id"`
				LabName     string `db:"lab_name"`
				SectionName string `db:"section_name"`
				CourseName  string `db:"course_name"`
				SectionID   string `db:"section_id"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[3] = fmt.Errorf("core labs scan: %w", err)
				return
			}
			out = append(out, models.CoreSearchLabResult{
				ID:          row.ID,
				LabName:     row.LabName,
				SectionName: row.SectionName,
				CourseName:  row.CourseName,
				Path:        "/sections/" + row.SectionID + "/labs/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Labs = out
		}
		mu.Unlock()
	}()

	// Enrolled section lab materials
	go func() {
		defer wg.Done()
		rows, err := r.db.QueryxContext(ctx, `
			SELECT
				m.id AS id,
				m.name AS material_name,
				l.display_name AS lab_name,
				s.name AS section_name,
				c.name AS course_name,
				s.id AS section_id,
				ls.lab_id AS lab_id
			FROM section_students ss
			JOIN sections s ON s.id = ss.section_id
			JOIN courses c ON c.id = s.course_id
			JOIN lab_sections ls ON ls.section_id = s.id
			JOIN labs l ON l.id = ls.lab_id
			JOIN lab_materials lm ON lm.lab_id = l.id
			JOIN materials m ON m.id = lm.material_id
			WHERE ss.student_id = $1
			  AND similarity(m.name, $2) > 0.1
			  AND ss.is_deleted = false
			  AND s.is_deleted = false
			  AND c.is_deleted = false
			  AND ls.is_deleted = false
			  AND l.is_deleted = false
			  AND lm.is_deleted = false
			  AND m.is_deleted = false
			ORDER BY similarity(m.name, $2) DESC
			LIMIT $3
		`, userID, q, limit)
		if err != nil {
			errs[4] = fmt.Errorf("core materials: %w", err)
			return
		}
		defer rows.Close()
		var out []models.CoreSearchMaterialResult
		for rows.Next() {
			var row struct {
				ID           string `db:"id"`
				MaterialName string `db:"material_name"`
				LabName      string `db:"lab_name"`
				SectionName  string `db:"section_name"`
				CourseName   string `db:"course_name"`
				SectionID    string `db:"section_id"`
				LabID        string `db:"lab_id"`
			}
			if err := rows.StructScan(&row); err != nil {
				errs[4] = fmt.Errorf("core materials scan: %w", err)
				return
			}
			out = append(out, models.CoreSearchMaterialResult{
				ID:           row.ID,
				MaterialName: row.MaterialName,
				LabName:      row.LabName,
				SectionName:  row.SectionName,
				CourseName:   row.CourseName,
				Path:         "/sections/" + row.SectionID + "/labs/" + row.LabID + "/materials/" + row.ID,
			})
		}
		mu.Lock()
		if out != nil {
			result.Materials = out
		}
		mu.Unlock()
	}()

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
