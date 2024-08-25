package main

import (
	"encoding/json"
	"io/ioutil"
)

type LocalizationSystem struct {
    translations map[string]map[string]string
}

func NewLocalizationSystem() *LocalizationSystem {
    return &LocalizationSystem{
        translations: make(map[string]map[string]string),
    }
}

func (ls *LocalizationSystem) LoadTranslations(language, filename string) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }

    var translations map[string]string
    err = json.Unmarshal(data, &translations)
    if err != nil {
        return err
    }

    ls.translations[language] = translations
    return nil
}

func (ls *LocalizationSystem) Translate(language, key string) string {
    if trans, ok := ls.translations[language]; ok {
        if str, ok := trans[key]; ok {
            return str
        }
    }
    return key
}
