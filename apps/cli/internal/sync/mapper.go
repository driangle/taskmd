package sync

// MappedTask holds the result of mapping an ExternalTask to taskmd fields.
type MappedTask struct {
	Title       string
	Description string
	Status      string
	Priority    string
	Owner       string
	Tags        []string
	URL         string
}

// MapExternalTask converts an ExternalTask to taskmd fields using the FieldMap.
func MapExternalTask(ext ExternalTask, fm FieldMap) MappedTask {
	m := MappedTask{
		Title:       ext.Title,
		Description: ext.Description,
		URL:         ext.URL,
	}

	m.Status = mapField(ext.Status, fm.Status, "pending")
	m.Priority = mapField(ext.Priority, fm.Priority, "")

	if fm.AssigneeToOwner {
		m.Owner = ext.Assignee
	}

	if fm.LabelsToTags {
		m.Tags = append(m.Tags, ext.Labels...)
	}

	return m
}

func mapField(value string, mapping map[string]string, fallback string) string {
	if mapping != nil {
		if mapped, ok := mapping[value]; ok {
			return mapped
		}
	}
	if value == "" {
		return fallback
	}
	return fallback
}
