// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pzsvc

import "time"

// DataDesc is the identifying information for a specific uploaded file
// or data block.  It is an important part of ingest requests, and the
// core of the response object for file metadata requests.
type DataDesc struct {
	DataID   string    `json:"dataId,omitempty"`
	DataType DataType  `json:"dataType,omitempty"`
	ResMeta  ResMeta   `json:"metadata,omitempty"`
	SpatMeta *SpatMeta `json:"spatialMetadata,omitempty"`
}

// DataType is highly polymorphic, and refers to a specific block of data.
// It defines the information specific to the type of that data block.
// Current options are 'body', 'geojson', 'literal', 'pointcloud', 'postgis',
// 'raster', 'shapefile', 'text' , 'urlparameter', and 'wfs'
type DataType struct { //name
	Content     string   `json:"content,omitempty"`           // the data itself, for those types small enough
	Type        string   `json:"type,omitempty"`              // see "current options", above
	MimeType    string   `json:"mimeType,omitempty"`          // mimetype corresponding to the type
	Location    *FileLoc `json:"location,omitempty"`          // where the data is, for those stored in shared folder/S3.
	DatTabName  string   `json:"databaseTableName,omitempty"` // geojson/shapefile, and only some of those.
	GeoJContent string   `json:"geoJsonContent,omitempty"`    // geojson, and only some of those.
	LitType     string   `json:"literalType,omitempty"`       // literal: DOUBLE,FLOAT,SHORT,LONG,BYTE,CHAR,BOOLEAN,STRING
	Database    string   `json:"database,omitempty"`          // postgis
	Table       string   `json:"table,omitempty"`             // postgis
	FeatureType string   `json:"featureType,omitempty"`       // wfs
	URL         string   `json:"url,omitempty"`               // wfs
	Version     string   `json:"version,omitempty"`           // wfs
}

// FileLoc contains information on where a file is located.  it's represented
// as an interface in the Pz code, and can be implemented in one fo two ways.
// Impl01 is the S3 bucket file store implementation.  Impl02 is the shared
// folder implementation.
type FileLoc struct {
	FileName   string `json:"fileName,omitempty"`
	FileSize   int    `json:"fileSize,omitempty"`
	Type       string `json:"type,omitempty"`       // "share" or "s3"
	BucketName string `json:"bucketName,omitempty"` // Impl01
	DomainName string `json:"domainName,omitempty"` // Impl01
	FilePath   string `json:"filePath,omitempty"`   // Impl02
}

// SpatMeta contains information on the spatial metadata (location) of a
// given image/feature/etc.  It is usually autogenerated
type SpatMeta struct {
	CoordRefSystem string  `json:"coordinateReferenceSystem,omitempty"`
	EpsgCode       int     `json:"epsgCode,omitempty"`
	MinX           float64 `json:"minX,omitempty"`
	MinY           float64 `json:"minY,omitempty"`
	MinZ           float64 `json:"minZ,omitempty"`
	MaxX           float64 `json:"maxX,omitempty"`
	MaxY           float64 `json:"maxY,omitempty"`
	MaxZ           float64 `json:"maxZ,omitempty"`
	NumFeatures    int     `json:"numFeatures,omitempty"`
}

// DeplStrct is an important part of the job status response for "deploy
// to geoserver" calls.
type DeplStrct struct {
	CapabilitiesURL string `json:"capabilitiesUrl,omitempty"`
	DataID          string `json:"dataId,omitempty"`
	DeplID          string `json:"deploymentId,omitempty"`
	Host            string `json:"host,omitempty"`
	Layer           string `json:"layer,omitempty"`
	Port            string `json:"port,omitempty"`
}

// DataResult is a hack to handle the fact that the backend we're addressing
// uses a lot of inheritence here.  Any one of five different classes could
// fill the slots set aside for DataResult objects in a job response.  Impl01
// is for ingest actions.  Impl02 is for deployment to geoserver.  Impl03 is
// for errors that occur doing one fo the others.  Impl04 and Impl05 are
// currently unknown.
type DataResult struct {
	DataID     string    `json:"dataId,omitempty"`     // Impl01
	Deployment DeplStrct `json:"deployment,omitempty"` // Impl02
	Details    string    `json:"details,omitempty"`    // Impl03
	Message    string    `json:"message,omitempty"`    // Impl03
	JobID      string    `json:"jobId,omitempty"`      // Impl04
	Text       string    `json:"text,omitempty"`       // Impl05
}

// JobProg is Pz's way of indicating job progress
type JobProg struct {
	PercentComplete int    `json:"percentComplete,omitempty"`
	TimeRemaining   string `json:"timeRemaining,omitempty"`
	TimeSpent       string `json:"timeSpent,omitempty"`
}

// ClassType holds classification data
type ClassType struct {
	Classification string `json:"classification,omitempty"`
}

// NumKeyVal is a String/Int Key/Value pair
type NumKeyVal struct {
	Key   string `json:"key,omitempty"`
	Value int    `json:"value,omitempty"`
}

// TxtKeyVal is a String/String Key/Value pair
type TxtKeyVal struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// PagStruct is the Pz pagination data format.  It is attached
// to all list-type structs.
type PagStruct struct {
	Count   int    `json:"count,omitempty"`
	Order   string `json:"order,omitempty"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"perPage,omitempty"`
	SortBy  string `json:"sortBy,omitempty"`
}

/***********************/
/*** Request Objects ***/
/***********************/

// Service is the Pz representation of a registered service.
// Used as the payload in register service and update service jobs.
// Also used in the response to the List Service job.
type Service struct {
	ContractURL string  `json:"contractUrl,omitempty"`
	Hearbeat    int     `json:"heartbeat,omitempty"`
	Method      string  `json:"method,omitempty"`
	ResMeta     ResMeta `json:"resourceMetadata,omitempty"`
	ServiceID   string  `json:"serviceId,omitempty"`
	Timeout     int     `json:"serviceId,timeout"`
	URL         string  `json:"url,omitempty"`
}

// IngestReq is the base object used to ingest a file to Piazza.
type IngestReq struct {
	Data DataDesc `json:"data,omitempty"`
	Host bool     `json:"host,omitempty"`
	Type string   `json:"type,omitempty"` // "ingest"
}

// ResMeta holds a resource medatadata. It's used broadly
// Worth noting that Pz pays no attention to the contents of the
// Metadata map field except to act as a passthrough.  Among
// other things, it's the request object for the "Change metadata"
// PUT request.
type ResMeta struct {
	Availability  string
	ClassType     ClassType         `json:"classType,omitempty"`
	CliCertReq    bool              `json:"clientCertRequired,omitempty"`
	Contacts      string            `json:"contacts,omitempty"`
	CreatedBy     string            `json:"createdBy,omitempty"`
	CreatedOn     string            `json:"createdOn,omitempty"`
	CredReq       bool              `json:"credentialsRequired,omitempty"`
	Description   string            `json:"description,omitempty"`
	Format        string            `json:"format,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Name          string            `json:"name,omitempty"`
	NetworkAvail  string            `json:"networkAvailable,omitempty"`
	NumKeyValList []NumKeyVal       `json:"numericKeyValueList,omitempty"`
	PreAuthReq    bool              `json:"preAuthRequired,omitempty"`
	QOS           string            `json:"qos,omitempty"`
	Reason        string            `json:"reason,omitempty"`
	StatusType    string            `json:"statusType,omitempty"`
	Tags          string            `json:"tags,omitempty"`
	TxtKeyValList []TxtKeyVal       `json:"textKeyValueList,omitempty"`
	Version       string            `json:"version,omitempty"`
}

/************************/
/*** Response Objects ***/
/************************/

// SvcList is the Pz representation of a list of service objects.
// It's the response object for a list/search services call
type SvcList struct {
	Type       string    `json:"type,omitempty"`
	Data       []Service `json:"data,omitempty"`
	Pagination PagStruct `json:"pagination,omitempty"`
}

// JobStatusResp is the response object to a Get Job Status call.
type JobStatusResp struct {
	CreatedBy string      `json:"createdBy,omitempty"`
	JobID     string      `json:"jobId,omitempty"`
	JobType   string      `json:"jobType,omitempty"`
	Progress  JobProg     `json:"progress,omitempty"`
	Result    *DataResult `json:"result,omitempty"`
	Status    string      `json:"status,omitempty"`
}

// JobInitResp is the immediate response object to all of the
// asynch "create job" calls (service calls and ingests, mostly).
// It can also be used as a request object for "repeat previously
// submitted job" calls
type JobInitResp struct {
	Data struct {
		JobID string `json:"jobId,omitempty"`
	} `json:"data,omitempty"`
}

// Alert is the response object to a Get Alert call when given
// an alertID.  It is also used to represent alerts in the search
// and list alerts calls
type Alert struct {
	AlertID   string `json:"alertId,omitempty"`
	CreatedBy string `json:"createdBy,omitempty"`
	EventID   string `json:"eventId,omitempty"`
	JobID     string `json:"jobId,omitempty"`
	TriggerID string `json:"triggerId,omitempty"`
}

// AlertList is the representation of a list of alert objects.
type AlertList struct {
	Type       string    `json:"type,omitempty"`
	Data       []Alert   `json:"data,omitempty"`
	Pagination PagStruct `json:"pagination,omitempty"`
}

// Trigger is the response object to a Get Trigger call when given
// a triggerID.  It is also used to represent triggers in the list
// and search trigger calls.  Additionally, it is used when creating
// or modifying a trigger.  Due to the complexity of the Trigger
// object, many of its component parts ahve been defined in a
// separate file.  Please see elasticSearchModel.go for those.
type Trigger struct {
	Name      string        `json:"name"`
	Enabled   bool          `json:"enabled"`
	Condition TrigCondition `json:"condition"`
	Job       TrigJob       `json:"job"`
	CreatedBy string        `json:"createdBy,omitempty"`
	CreatedOn string        `json:"createdOn,omitempty"`
	TriggerID string        `json:"triggerId,omitempty"`
}

// TriggerList ...
type TriggerList struct {
	Type       string    `json:"type,omitempty"`
	Data       []Trigger `json:"data,omitempty"`
	Pagination PagStruct `json:"pagination,omitempty"`
}

// EventType ...
type EventType struct {
	EventTypeID string            `json:"eventTypeId"`
	Name        string            `json:"name" binding:"required"`
	Mapping     map[string]string `json:"mapping" binding:"required"`
	CreatedBy   string            `json:"createdBy"`
	CreatedOn   time.Time         `json:"createdOn"`
}

// EventTypeList ...
type EventTypeList struct {
	Type       string      `json:"type,omitempty"`
	Data       []EventType `json:"data,omitempty"`
	Pagination PagStruct   `json:"pagination,omitempty"`
}

// Event ...
type Event struct {
	EventID      string                 `json:"eventId"`
	EventTypeID  string                 `json:"eventTypeId" binding:"required"`
	Data         map[string]interface{} `json:"data"`
	CreatedBy    string                 `json:"createdBy"`
	CreatedOn    time.Time              `json:"createdOn"`
	CronSchedule string                 `json:"cronSchedule"`
}

// EventList ...
type EventList struct {
	Type       string    `json:"type,omitempty"`
	Data       []Event   `json:"data,omitempty"`
	Pagination PagStruct `json:"pagination,omitempty"`
}
