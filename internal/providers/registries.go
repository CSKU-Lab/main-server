package providers

import (
	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/registrables"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"

	"github.com/google/wire"
)

func ProvideTypingSubmission(
	repo repositories.TypingSubmissionRepository,
	typingMatRepo repositories.TypingMaterialRepository,
	cfg *configs.Config,
) *registrables.TypingSubmission {
	return registrables.NewTypingSubmission(repo, typingMatRepo, cfg.JWTSecret)
}

func NewPopulatedMaterialRegistry(
	codeMat *registrables.CodeMaterial,
	typingMat *registrables.TypingMaterial,
	docMat *registrables.DocumentMaterial,
) registries.Material {
	r := registries.NewMaterialRegistry()
	r.Register("code", codeMat)
	r.Register("typing", typingMat)
	r.Register("document", docMat)
	return r
}

func NewPopulatedSubmissionRegistry(
	code *registrables.CodeSubmission,
	typing *registrables.TypingSubmission,
	document *registrables.DocumentSubmission,
) registries.SubmissionRegistry {
	r := registries.NewSubmission()
	r.Register("code", code)
	r.Register("typing", typing)
	r.Register("document", document)
	return r
}

func NewPopulatedAffectedEntityFactory(
	course *registrables.DeletedCourseAffected,
	semester *registrables.DeletedSemesterAffected,
	section *registrables.DeletedSectionAffected,
	lab *registrables.DeletedLabAffected,
	labSection *registrables.DeletedLabSectionAffected,
) registries.AffectedEntitiesFactory {
	f := registries.NewAffectedEntityFactory()
	f.Register("course", course)
	f.Register("semester", semester)
	f.Register("section", section)
	f.Register("lab", lab)
	f.Register("lab_section", labSection)
	return f
}

var RegistrySet = wire.NewSet(
	NewPopulatedMaterialRegistry,
	NewPopulatedSubmissionRegistry,
	NewPopulatedAffectedEntityFactory,
	registrables.NewCodeMaterial,
	registrables.NewTypingMaterial,
	registrables.NewDocumentMaterial,
	registrables.NewCodeSubmission,
	ProvideTypingSubmission,
	registrables.NewDocumentSubmission,
	registrables.NewDeletedCourseAffected,
	registrables.NewDeletedSemesterAffected,
	registrables.NewDeletedSectionAffected,
	registrables.NewDeletedLabAffected,
	registrables.NewDeletedLabSectionAffected,
)
