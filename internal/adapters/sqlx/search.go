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
