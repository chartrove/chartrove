package database

import (
	"context"
	"fmt"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoadSystemSettings will retrieve settings from the database, and return a SystemSettings struct
func (gr *GormRepository) LoadSystemSettings(ctx context.Context) (*models.SystemSettings, error) {

	settingsEntries := []models.SystemSettingEntry{}
	if err := gr.GormClient.
		WithContext(ctx).
		Find(&settingsEntries).Error; err != nil {
		return nil, fmt.Errorf("Could not get settings from DB: %v", err)
	}

	settings := models.SystemSettings{}
	for _, settingsEntry := range settingsEntries {
		err := settings.FromSystemSettingsEntry(&settingsEntry)
		if err != nil {
			return nil, fmt.Errorf("Could not get settings from DB: %v", err)
		}
	}

	return &settings, nil
}

// testing
// SaveSystemSettings will update save the settings to the database.
func (gr *GormRepository) SaveSystemSettings(ctx context.Context, newSettings *models.SystemSettings) error {

	// Installation ID and secret are one credential pair. A later row failure
	// must roll back earlier writes and propagate the actual storage error.
	return gr.GormClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		currentSettingsEntries := []models.SystemSettingEntry{}
		if err := tx.Where("setting_key_name IN ?", []string{"installation_id", "installation_secret"}).Find(&currentSettingsEntries).Error; err != nil {
			return fmt.Errorf("could not get system settings from DB: %w", err)
		}
		if len(currentSettingsEntries) != 2 {
			return fmt.Errorf("system settings credential rows unavailable")
		}
		seen := make(map[string]bool, 2)
		for _, entry := range currentSettingsEntries {
			if entry.SettingDataType != "string" || seen[entry.SettingKeyName] {
				return fmt.Errorf("system settings credential rows invalid")
			}
			seen[entry.SettingKeyName] = true
		}
		newSettingsEntries, err := newSettings.ToSystemSettingsEntry(currentSettingsEntries)
		if err != nil {
			return fmt.Errorf("merge new system settings with DB: %w", err)
		}
		for ndx, settingsEntry := range newSettingsEntries {
			var result *gorm.DB
			if settingsEntry.ID == uuid.Nil {
				result = tx.Model(&models.SystemSettingEntry{}).Create(&settingsEntry)
			} else {
				result = tx.Model(&models.SystemSettingEntry{}).
					Where([]uuid.UUID{settingsEntry.ID}).
					Select("setting_value_numeric", "setting_value_string", "setting_value_bool", "setting_value_array").
					Updates(newSettingsEntries[ndx])
			}
			if result.Error != nil {
				return fmt.Errorf("could not save system settings: %w", result.Error)
			}
			if result.RowsAffected != 1 {
				return fmt.Errorf("system settings credential row update incomplete")
			}
		}
		return nil
	})
}

// LoadSettings will retrieve settings from the database, store them in the AppConfig object, and return a Settings struct
func (gr *GormRepository) LoadUserSettings(ctx context.Context) (*models.UserSettings, error) {
	currentUser, currentUserErr := gr.GetCurrentUser(ctx)
	if currentUserErr != nil {
		return nil, currentUserErr
	}

	settingsEntries := []models.UserSettingEntry{}
	if err := gr.GormClient.
		WithContext(ctx).
		Where(models.UserSettingEntry{
			UserID: currentUser.ID,
		}).
		Find(&settingsEntries).Error; err != nil {
		return nil, fmt.Errorf("Could not get settings from DB: %v", err)
	}

	settings := models.UserSettings{}
	for _, settingsEntry := range settingsEntries {
		err := settings.FromUserSettingsEntry(&settingsEntry)
		if err != nil {
			return nil, fmt.Errorf("Could not get settings from DB: %v", err)
		}
	}

	return &settings, nil
}

// testing
// curl -d '{"metrics": { "notify_level": 5, "status_filter_attributes": 5, "status_threshold": 5 }}' -H "Content-Type: application/json" -X POST http://localhost:9090/api/settings
// SaveSettings will update settings in AppConfig object, then save the settings to the database.
func (gr *GormRepository) SaveUserSettings(ctx context.Context, newSettings *models.UserSettings) error {
	currentUser, currentUserErr := gr.GetCurrentUser(ctx)
	if currentUserErr != nil {
		return currentUserErr
	}

	//retrieve current settings from the database
	currentSettingsEntries := []models.UserSettingEntry{}

	if err := gr.GormClient.
		WithContext(ctx).
		Where(models.UserSettingEntry{
			UserID: currentUser.ID,
		}).
		Find(&currentSettingsEntries).Error; err != nil {
		return fmt.Errorf("Could not get settings from DB: %v", err)
	}

	//update settingsEntries

	newSettingsEntries, err := newSettings.ToUserSettingsEntry(currentSettingsEntries)
	if err != nil {
		return fmt.Errorf("merge new settings with DB: %v", err)
	}

	for ndx, settingsEntry := range newSettingsEntries {

		// store in database.
		//TODO: this should be `gr.gormClient.Updates(&settingsEntries).Error`
		err := gr.GormClient.
			WithContext(ctx).
			Model(&models.UserSettingEntry{}).
			Where([]uuid.UUID{settingsEntry.ID}).
			Select("setting_value_numeric", "setting_value_string", "setting_value_bool", "setting_value_array").
			Updates(newSettingsEntries[ndx]).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (gr *GormRepository) PopulateDefaultUserSettings(ctx context.Context, userId uuid.UUID) error {

	//retrieve current settings from the database
	settingsEntries := []models.UserSettingEntry{}
	settingsEntries = append(settingsEntries, models.UserSettingEntry{
		UserID:                userId,
		SettingKeyName:        "dashboard_locations",
		SettingKeyDescription: "remote dashboard locations (github gists)",
		SettingDataType:       "array",
		SettingValueArray:     []string{},
	})

	return gr.GormClient.WithContext(ctx).Create(settingsEntries).Error

}
