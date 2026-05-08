package i18n

type Localizer struct {
	lang string
}

func New(lang string) Localizer {
	if lang != "zh_cn" {
		lang = "en"
	}
	return Localizer{lang: lang}
}

func (l Localizer) T(key string) string {
	if l.lang == "zh_cn" {
		if v, ok := zhCN[key]; ok {
			return v
		}
	}
	if v, ok := en[key]; ok {
		return v
	}
	return key
}

func (l Localizer) Lang() string {
	return l.lang
}
