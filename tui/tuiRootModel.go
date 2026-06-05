package tui

import (
	bbt "charm.land/bubbletea/v2"
	//intTea "charm.land/bubbletea/v2/tea"
)

type welcomeModel struct {
	mWidth uint32
	mHeight uint32 
	mChoice int
}

func(w welcomeModel) Init() bbt.Cmd{
	return bbt.ClearScreen
}

func(w welcomeModel) Update( msg bbt.Msg) (bbt.Model,bbt.Cmd){
	return w,bbt.Quit
}

func(w welcomeModel) View() bbt.View {
	return bbt.View{}
}


func CreateModel (modelName string) (bbt.Model,bool) {
	switch modelName {
		case "welcomeModel": return welcomeModel{}, true
	default: return nil,false
	}

}