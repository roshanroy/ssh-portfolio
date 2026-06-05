package tui

import (
	bbt "charm.land/bubbletea/v2"
	//intTea "charm.land/bubbletea/v2/tea"
)



func CreateModel (modelName string) (bbt.Model,bool) {
	switch modelName {
		case "welcomeModel": return welcomeModel{}, true
	default: return nil,false
	}

}