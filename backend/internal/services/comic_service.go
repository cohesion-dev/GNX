package services

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend/internal/models"
	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/internal/utils"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type ComicService struct {
	comicRepo      *repositories.ComicRepository
	roleRepo       *repositories.RoleRepository
	sectionRepo    *repositories.SectionRepository
	storyboardRepo *repositories.StoryboardRepository
	storageService *storage.QiniuClient
	aigcService    *gnxaigc.GnxAIGC
	db             *gorm.DB
}

func NewComicService(
	comicRepo *repositories.ComicRepository,
	roleRepo *repositories.RoleRepository,
	sectionRepo *repositories.SectionRepository,
	storyboardRepo *repositories.StoryboardRepository,
	storageService *storage.QiniuClient,
	aigcService *gnxaigc.GnxAIGC,
	db *gorm.DB,
) *ComicService {
	return &ComicService{
		comicRepo:      comicRepo,
		roleRepo:       roleRepo,
		sectionRepo:    sectionRepo,
		storyboardRepo: storyboardRepo,
		storageService: storageService,
		aigcService:    aigcService,
		db:             db,
	}
}

func (s *ComicService) GetComicList(page, limit int, status string) ([]models.Comic, int64, error) {
	offset := (page - 1) * limit
	return s.comicRepo.GetListWithFilter(limit, offset, status)
}

func (s *ComicService) CreateComic(title, userPrompt string, fileContent io.Reader) (*models.Comic, error) {
	existingComic, err := s.comicRepo.GetByTitle(title)
	if err == nil && existingComic.ID > 0 {
		return nil, fmt.Errorf("comic with title '%s' already exists", title)
	}

	content, err := io.ReadAll(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	comic := &models.Comic{
		Title:          title,
		UserPrompt:     userPrompt,
		Status:         "pending",
		HasMoreContent: true,
	}

	if err := s.comicRepo.Create(comic); err != nil {
		return nil, fmt.Errorf("failed to create comic: %w", err)
	}

	go s.processComicGeneration(comic.ID, content)

	return comic, nil
}

func (s *ComicService) AppendSections(comicID uint, fileContent io.Reader) error {
	comic, err := s.comicRepo.GetByIDWithRelations(comicID)
	if err != nil {
		return fmt.Errorf("comic not found: %w", err)
	}

	content, err := io.ReadAll(fileContent)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	go s.processSectionAppend(comic, content)

	return nil
}

func (s *ComicService) GetComicDetail(id uint) (*models.Comic, error) {
	return s.comicRepo.GetByIDWithRelations(id)
}

func (s *ComicService) GetComicSections(id uint) ([]models.ComicSection, error) {
	return s.sectionRepo.GetByComicID(id)
}

func (s *ComicService) processComicGeneration(comicID uint, content []byte) {
	defer func() {
		if r := recover(); r != nil {
			s.comicRepo.UpdateStatus(comicID, "failed")
		}
	}()

	s.comicRepo.UpdateStatus(comicID, "processing")

	novelFileKey := fmt.Sprintf("comics/%d/novel_%d.txt", comicID, time.Now().Unix())
	novelFileURL, err := s.storageService.UploadFile(novelFileKey, content, "text/plain")
	if err != nil {
		s.comicRepo.UpdateStatus(comicID, "failed")
		return
	}

	if err := s.comicRepo.UpdateNovelFileURL(comicID, novelFileURL); err != nil {
		s.comicRepo.UpdateStatus(comicID, "failed")
		return
	}

	comic, err := s.comicRepo.GetByIDWithRelations(comicID)
	if err != nil {
		s.comicRepo.UpdateStatus(comicID, "failed")
		return
	}

	ctx := context.Background()
	metadata, err := s.aigcService.GenerateComicMetadata(ctx, gnxaigc.ComicMetadataInput{
		NovelTitle:   comic.Title,
		NovelContent: string(content),
		UserPrompt:   comic.UserPrompt,
	})
	if err == nil {
		updates := map[string]interface{}{
			"brief":       metadata.Brief,
			"icon_prompt": metadata.IconPrompt,
			"bg_prompt":   metadata.BgPrompt,
		}
		s.comicRepo.UpdateComicInfo(comicID, updates)
	}

	sections := utils.ParseNovelSections(string(content))

	for _, section := range sections {
		if err := s.processSection(comic.ID, section); err != nil {
			continue
		}
	}

	s.generateComicIconAndBg(comicID)

	s.comicRepo.UpdateStatus(comicID, "completed")
}

func (s *ComicService) processSectionAppend(comic *models.Comic, content []byte) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	sections := utils.ParseNovelSections(string(content))

	for _, section := range sections {
		if err := s.processSection(comic.ID, section); err != nil {
			continue
		}
	}

	if !comic.HasMoreContent && comic.Icon == "" {
		s.generateComicIconAndBg(comic.ID)
	}
}

func (s *ComicService) processSection(comicID uint, section utils.Section) error {
	comic, err := s.comicRepo.GetByIDWithRelations(comicID)
	if err != nil {
		return err
	}

	comicSection := &models.ComicSection{
		ComicID: comicID,
		Index:   section.Index,
		Title:   section.Title,
		Detail:  section.Content,
		Status:  "processing",
	}

	if err := s.sectionRepo.Create(comicSection); err != nil {
		return err
	}

	ctx := context.Background()

	availableVoices, _ := s.aigcService.GetVoiceList(ctx)
	var ttsVoices []gnxaigc.TTSVoiceItem
	for _, v := range availableVoices {
		ttsVoices = append(ttsVoices, gnxaigc.TTSVoiceItem{
			VoiceName: v.VoiceName,
			VoiceType: v.VoiceType,
		})
	}

	var characterFeatures []gnxaigc.CharacterFeature
	for _, role := range comic.Roles {
		characterFeatures = append(characterFeatures, gnxaigc.CharacterFeature{
			Basic: gnxaigc.CharacterBasicProfile{
				Name:   role.Name,
				Gender: role.Gender,
				Age:    role.Age,
			},
			Visual: gnxaigc.CharacterVisualProfile{
				Hair:               role.Hair,
				HabitualExpression: role.HabitualExpr,
				SkinTone:           role.SkinTone,
				FaceShape:          role.FaceShape,
			},
			TTS: gnxaigc.CharacterTTSProfile{
				VoiceName:  role.VoiceName,
				VoiceType:  role.VoiceType,
				SpeedRatio: role.SpeedRatio,
			},
			Comment: role.Brief,
		})
	}

	input := gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         section.Title,
		Content:              section.Content,
		AvailableVoiceStyles: ttsVoices,
		CharacterFeatures:    characterFeatures,
	}

	output, err := s.aigcService.SummaryChapter(ctx, input)
	if err != nil {
		s.sectionRepo.UpdateStatus(comicSection.ID, "failed")
		return err
	}

	roleMap := s.updateCharacterFeatures(comicID, output.CharacterFeatures)

	for pageIdx, aiPage := range output.StoryboardPages {
		page := &models.ComicStoryboardPage{
			SectionID:   comicSection.ID,
			Index:       pageIdx + 1,
			ImagePrompt: aiPage.ImagePrompt,
			LayoutHint:  aiPage.LayoutHint,
			PageSummary: aiPage.PageSummary,
			Status:      "pending",
		}

		if err := s.storyboardRepo.CreatePage(page); err != nil {
			continue
		}

		for panelIdx, aiPanel := range aiPage.Panels {
			panel := &models.ComicStoryboardPanel{
				SectionID:    comicSection.ID,
				PageID:       page.ID,
				Index:        panelIdx + 1,
				VisualPrompt: aiPanel.VisualPrompt,
				PanelSummary: aiPanel.PanelSummary,
				Status:       "pending",
			}

			if err := s.storyboardRepo.CreatePanel(panel); err != nil {
				continue
			}

			for segIdx, aiSegment := range aiPanel.SourceTextSegments {
				var roleID *uint
				if len(aiSegment.CharacterNames) > 0 {
					characterName := aiSegment.CharacterNames[0]
					if id, exists := roleMap[characterName]; exists {
						roleID = &id
					}
				}

				segment := &models.ComicStoryboardSegment{
					PanelID:       panel.ID,
					Index:         segIdx + 1,
					Text:          aiSegment.Text,
					CharacterRefs: aiSegment.CharacterNames,
					RoleID:        roleID,
				}

				if err := s.storyboardRepo.CreateSegment(segment); err != nil {
					continue
				}
			}

			go s.generatePanelImage(panel.ID, aiPanel)
		}

	}

	s.sectionRepo.UpdateStatus(comicSection.ID, "completed")
	return nil
}

func (s *ComicService) updateCharacterFeatures(comicID uint, features []gnxaigc.CharacterFeature) map[string]uint {
	roleMap := make(map[string]uint)

	for _, feature := range features {
		existingRole, err := s.roleRepo.GetByNameAndComicID(feature.Basic.Name, comicID)
		if err != nil || existingRole == nil {
			role := &models.ComicRole{
				ComicID:      comicID,
				Name:         feature.Basic.Name,
				Gender:       feature.Basic.Gender,
				Age:          feature.Basic.Age,
				Hair:         feature.Visual.Hair,
				HabitualExpr: feature.Visual.HabitualExpression,
				SkinTone:     feature.Visual.SkinTone,
				FaceShape:    feature.Visual.FaceShape,
				VoiceName:    feature.TTS.VoiceName,
				VoiceType:    feature.TTS.VoiceType,
				SpeedRatio:   feature.TTS.SpeedRatio,
				Brief:        feature.Comment,
			}
			if err := s.roleRepo.Create(role); err == nil {
				roleMap[role.Name] = role.ID
			}
		} else {
			updates := map[string]interface{}{
				"gender":        feature.Basic.Gender,
				"age":           feature.Basic.Age,
				"hair":          feature.Visual.Hair,
				"habitual_expr": feature.Visual.HabitualExpression,
				"skin_tone":     feature.Visual.SkinTone,
				"face_shape":    feature.Visual.FaceShape,
				"voice_name":    feature.TTS.VoiceName,
				"voice_type":    feature.TTS.VoiceType,
				"speed_ratio":   feature.TTS.SpeedRatio,
				"brief":         feature.Comment,
			}
			s.roleRepo.UpdateByID(existingRole.ID, updates)
			roleMap[existingRole.Name] = existingRole.ID
		}
	}

	return roleMap
}

func (s *ComicService) generatePanelImage(panelID uint, aiPanel gnxaigc.StoryboardPanel) {
	ctx := context.Background()

	imageData, err := s.aigcService.GenerateImageByText(ctx, aiPanel.VisualPrompt)
	if err != nil {
		s.storyboardRepo.UpdatePanelStatus(panelID, "failed")
		return
	}

	imageKey := fmt.Sprintf("panels/%d_%d.png", panelID, time.Now().Unix())
	imageURL, err := s.storageService.UploadImage(imageKey, imageData)
	if err != nil {
		s.storyboardRepo.UpdatePanelStatus(panelID, "failed")
		return
	}

	s.storyboardRepo.UpdatePanelImageURL(panelID, imageURL)
	s.storyboardRepo.UpdatePanelStatus(panelID, "completed")
}

func (s *ComicService) generateComicIconAndBg(comicID uint) {
	ctx := context.Background()
	comic, err := s.comicRepo.GetByIDWithRelations(comicID)
	if err != nil {
		return
	}

	var wg sync.WaitGroup

	if comic.Icon == "" && comic.IconPrompt != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			iconData, err := s.aigcService.GenerateImageByText(ctx, comic.IconPrompt)
			if err == nil {
				iconKey := fmt.Sprintf("comics/%d/icon_%d.png", comicID, time.Now().Unix())
				iconURL, err := s.storageService.UploadImage(iconKey, iconData)
				if err == nil {
					s.comicRepo.UpdateIcon(comicID, iconURL)
				}
			}
		}()
	}

	if comic.Bg == "" && comic.BgPrompt != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bgData, err := s.aigcService.GenerateImageByText(ctx, comic.BgPrompt)
			if err == nil {
				bgKey := fmt.Sprintf("comics/%d/bg_%d.png", comicID, time.Now().Unix())
				bgURL, err := s.storageService.UploadImage(bgKey, bgData)
				if err == nil {
					s.comicRepo.UpdateBg(comicID, bgURL)
				}
			}
		}()
	}

	wg.Wait()
}
