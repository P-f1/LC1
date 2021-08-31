/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package execlistener

import (
	"context"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/P-f1/LC/labs-flogo-lib/exec/execeventbroker"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

const (
	cConnection     = "execConnection"
	cConnectionName = "name"
)

//-============================================-//
//   Entry point register Trigger & factory
//-============================================-//

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&ExecListener{}, &Factory{})
}

//-===============================-//
//     Define Trigger Factory
//-===============================-//

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	settings := &Settings{}
	err := metadata.MapToStruct(config.Settings, settings, true)
	if err != nil {
		return nil, err
	}

	return &ExecListener{settings: settings}, nil
}

//-=========================-//
//      Define Trigger
//-=========================-//

var logger log.Logger

type ExecListener struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	server   *execeventbroker.EXEEventBroker
	mux      sync.Mutex

	settings *Settings
	handlers []trigger.Handler
}

// implements trigger.Initializable.Initialize
func (this *ExecListener) Initialize(ctx trigger.InitContext) error {

	this.handlers = ctx.GetHandlers()
	logger = ctx.Logger()

	return nil
}

// implements ext.Trigger.Start
func (this *ExecListener) Start() error {

	logger.Info("Start")
	handlers := this.handlers

	logger.Info("Processing handlers")

	connection, exist := handlers[0].Settings()[cConnection]
	if !exist {
		return activity.NewError("SSE connection is not configured", "TGDB-SSE-4001", nil)
	}

	connectionInfo, _ := data.CoerceToObject(connection)
	if connectionInfo == nil {
		return activity.NewError("SSE connection not able to be parsed", "TGDB-SSE-4002", nil)
	}

	var serverId string
	connectionSettings, _ := connectionInfo["settings"].([]interface{})
	if connectionSettings != nil {
		for _, v := range connectionSettings {
			setting, err := data.CoerceToObject(v)
			if nil != err {
				continue
			}

			if nil != setting {
				if setting["name"] == cConnectionName {
					serverId = setting["value"].(string)
				}
			}

		}

		var err error
		this.server, err = execeventbroker.GetFactory().CreateEXEEventBroker(serverId, this)
		if nil != err {
			return err
		}
		logger.Info("Server = ", *this.server)
		go this.server.Start()
	}

	return nil
}

// implements ext.Trigger.Stop
func (this *ExecListener) Stop() error {
	this.server.Stop()
	return nil
}

func (this *ExecListener) ProcessEvent(event map[string]interface{}) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	logger.Debug("Got Exec event : ", event)
	outputData := &Output{}
	outputData.Event = event
	logger.Debug("Send Exec event out : ", outputData)

	_, err := this.handlers[0].Handle(context.Background(), outputData)
	if nil != err {
		logger.Info("Error -> ", err)
	}

	return err
}
