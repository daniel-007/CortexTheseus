// Copyright (c) 2013-2014, Jeffrey Wilcke. All rights reserved.
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either
// version 2.1 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this library; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
// MA 02110-1301  USA

package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/ethchain"
	"github.com/ethereum/go-ethereum/ethlog"
	"github.com/ethereum/go-ethereum/ethpipe"
	"github.com/ethereum/go-ethereum/ethutil"
	"github.com/ethereum/go-ethereum/utils"
)

type plugin struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// LogPrint writes to the GUI log.
func (gui *Gui) LogPrint(level ethlog.LogLevel, msg string) {
	/*
		str := strings.TrimRight(s, "\n")
		lines := strings.Split(str, "\n")

		view := gui.getObjectByName("infoView")
		for _, line := range lines {
			view.Call("addLog", line)
		}
	*/
}
func (gui *Gui) Transact(recipient, value, gas, gasPrice, d string) (*ethpipe.JSReceipt, error) {
	var data string
	if len(recipient) == 0 {
		code, err := ethutil.Compile(d, false)
		if err != nil {
			return nil, err
		}
		data = ethutil.Bytes2Hex(code)
	} else {
		data = ethutil.Bytes2Hex(utils.FormatTransactionData(d))
	}

	return gui.pipe.Transact(gui.privateKey(), recipient, value, gas, gasPrice, data)
}

func (gui *Gui) SetCustomIdentifier(customIdentifier string) {
	gui.clientIdentity.SetCustomIdentifier(customIdentifier)
	gui.config.Save("id", customIdentifier)
}

func (gui *Gui) GetCustomIdentifier() string {
	return gui.clientIdentity.GetCustomIdentifier()
}

func (gui *Gui) ToggleTurboMining() {
	gui.miner.ToggleTurbo()
}

// functions that allow Gui to implement interface ethlog.LogSystem
func (gui *Gui) SetLogLevel(level ethlog.LogLevel) {
	gui.logLevel = level
	gui.stdLog.SetLogLevel(level)
	gui.config.Save("loglevel", level)
}

func (gui *Gui) GetLogLevel() ethlog.LogLevel {
	return gui.logLevel
}

func (self *Gui) AddPlugin(pluginPath string) {
	self.plugins[pluginPath] = plugin{Name: pluginPath, Path: pluginPath}

	json, _ := json.MarshalIndent(self.plugins, "", "    ")
	ethutil.WriteFile(ethutil.Config.ExecPath+"/plugins.json", json)
}

func (self *Gui) RemovePlugin(pluginPath string) {
	delete(self.plugins, pluginPath)

	json, _ := json.MarshalIndent(self.plugins, "", "    ")
	ethutil.WriteFile(ethutil.Config.ExecPath+"/plugins.json", json)
}

// this extra function needed to give int typecast value to gui widget
// that sets initial loglevel to default
func (gui *Gui) GetLogLevelInt() int {
	return int(gui.logLevel)
}
func (self *Gui) DumpState(hash, path string) {
	var stateDump []byte

	if len(hash) == 0 {
		stateDump = self.eth.StateManager().CurrentState().Dump()
	} else {
		var block *ethchain.Block
		if hash[0] == '#' {
			i, _ := strconv.Atoi(hash[1:])
			block = self.eth.ChainManager().GetBlockByNumber(uint64(i))
		} else {
			block = self.eth.ChainManager().GetBlock(ethutil.Hex2Bytes(hash))
		}

		if block == nil {
			logger.Infof("block err: not found %s\n", hash)
			return
		}

		stateDump = block.State().Dump()
	}

	file, err := os.OpenFile(path[7:], os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		logger.Infoln("dump err: ", err)
		return
	}
	defer file.Close()

	logger.Infof("dumped state (%s) to %s\n", hash, path)

	file.Write(stateDump)
}
func (gui *Gui) ToggleMining() {
	var txt string
	if gui.eth.Mining {
		utils.StopMining(gui.eth)
		txt = "Start mining"

		gui.getObjectByName("miningLabel").Set("visible", false)
	} else {
		utils.StartMining(gui.eth)
		gui.miner = utils.GetMiner()
		txt = "Stop mining"

		gui.getObjectByName("miningLabel").Set("visible", true)
	}

	gui.win.Root().Set("miningButtonText", txt)
}
