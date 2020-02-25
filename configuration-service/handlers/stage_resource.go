package handlers

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/keptn/go-utils/pkg/utils"
	"github.com/keptn/keptn/configuration-service/common"
	"github.com/keptn/keptn/configuration-service/config"
	"github.com/keptn/keptn/configuration-service/models"
	"github.com/keptn/keptn/configuration-service/restapi/operations/stage_resource"
)

// GetProjectProjectNameStageStageNameResourceHandlerFunc get list of stage resources
func GetProjectProjectNameStageStageNameResourceHandlerFunc(params stage_resource.GetProjectProjectNameStageStageNameResourceParams) middleware.Responder {
	common.LockProject(params.ProjectName)
	defer common.UnlockProject(params.ProjectName)
	logger := utils.NewLogger("", "", "configuration-service")
	if !common.StageExists(params.ProjectName, params.StageName) {
		return stage_resource.NewGetProjectProjectNameStageStageNameResourceNotFound().WithPayload(&models.Error{Code: 404, Message: swag.String("Stage does not exist")})
	}
	logger.Debug("Checking out " + params.StageName + " branch")
	err := common.CheckoutBranch(params.ProjectName, params.StageName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewGetProjectProjectNameStageStageNameResourceDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String("Could not retrieve stage resources")})
	}

	projectConfigPath := config.ConfigDir + "/" + params.ProjectName
	result := common.GetPaginatedResources(projectConfigPath, params.PageSize, params.NextPageKey)
	return stage_resource.NewGetProjectProjectNameStageStageNameResourceOK().WithPayload(result)
}

// GetProjectProjectNameStageStageNameResourceResourceURIHandlerFunc get the specified resource
func GetProjectProjectNameStageStageNameResourceResourceURIHandlerFunc(params stage_resource.GetProjectProjectNameStageStageNameResourceResourceURIParams) middleware.Responder {
	common.LockProject(params.ProjectName)
	defer common.UnlockProject(params.ProjectName)
	logger := utils.NewLogger("", "", "configuration-service")
	projectConfigPath := config.ConfigDir + "/" + params.ProjectName
	resourcePath := projectConfigPath + "/" + params.ResourceURI
	if !common.StageExists(params.ProjectName, params.StageName) {
		return stage_resource.NewGetProjectProjectNameStageStageNameResourceResourceURINotFound().WithPayload(&models.Error{Code: 404, Message: swag.String("Project not found")})
	}
	logger.Debug("Checking out " + params.StageName + " branch")
	err := common.CheckoutBranch(params.ProjectName, params.StageName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewGetProjectProjectNameStageStageNameResourceResourceURIDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String("Could not check out branch containing stage config")})
	}

	if !common.FileExists(resourcePath) {
		return stage_resource.NewGetProjectProjectNameStageStageNameResourceResourceURINotFound().WithPayload(&models.Error{Code: 404, Message: swag.String("Stage resource not found")})
	}

	dat, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewGetProjectProjectNameStageStageNameResourceResourceURIDefault(500).WithPayload(&models.Error{Code: 500, Message: swag.String("Could not read file")})
	}

	resourceContent := base64.StdEncoding.EncodeToString(dat)
	return stage_resource.NewGetProjectProjectNameStageStageNameResourceResourceURIOK().WithPayload(
		&models.Resource{
			ResourceURI:     &params.ResourceURI,
			ResourceContent: resourceContent,
		})
}

// PostProjectProjectNameStageStageNameResourceHandlerFunc creates list of new resources in a stage
func PostProjectProjectNameStageStageNameResourceHandlerFunc(params stage_resource.PostProjectProjectNameStageStageNameResourceParams) middleware.Responder {
	common.LockProject(params.ProjectName)
	defer common.UnlockProject(params.ProjectName)
	logger := utils.NewLogger("", "", "configuration-service")
	if !common.StageExists(params.ProjectName, params.StageName) {
		return stage_resource.NewPostProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Stage does not exist")})
	}

	projectConfigPath := config.ConfigDir + "/" + params.ProjectName

	logger.Debug("Creating new resource(s) in: " + projectConfigPath + " in stage " + params.StageName)
	logger.Debug("Checking out branch: " + params.StageName)
	err := common.CheckoutBranch(params.ProjectName, params.StageName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPostProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not check out branch containing stage config")})
	}

	for _, res := range params.Resources.Resources {
		filePath := projectConfigPath + "/" + *res.ResourceURI
		logger.Debug("Adding resource: " + filePath)
		common.WriteBase64EncodedFile(filePath, res.ResourceContent)
	}

	logger.Debug("Staging Changes")
	err = common.StageAndCommitAll(params.ProjectName, "Added resources")
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPostProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not commit changes")})
	}
	logger.Debug("Successfully added resources")

	newVersion, err := common.GetCurrentVersion(params.ProjectName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPostProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not retrieve latest vesion")})
	}
	return stage_resource.NewPostProjectProjectNameStageStageNameResourceCreated().WithPayload(&models.Version{
		Version: newVersion,
	})
}

// PutProjectProjectNameStageStageNameResourceHandlerFunc updates list of stage resources
func PutProjectProjectNameStageStageNameResourceHandlerFunc(params stage_resource.PutProjectProjectNameStageStageNameResourceParams) middleware.Responder {
	common.LockProject(params.ProjectName)
	defer common.UnlockProject(params.ProjectName)
	logger := utils.NewLogger("", "", "configuration-service")
	if !common.StageExists(params.ProjectName, params.StageName) {
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Stage does not exist")})
	}
	projectConfigPath := config.ConfigDir + "/" + params.ProjectName

	logger.Debug("Creating new resource(s) in: " + projectConfigPath + " in stage " + params.StageName)
	logger.Debug("Checking out branch: " + params.StageName)
	err := common.CheckoutBranch(params.ProjectName, params.StageName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not check out branch")})
	}

	for _, res := range params.Resources.Resources {
		filePath := projectConfigPath + "/" + *res.ResourceURI
		common.WriteBase64EncodedFile(filePath, res.ResourceContent)
	}

	logger.Debug("Staging Changes")
	err = common.StageAndCommitAll(params.ProjectName, "Updated resources")
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not commit changes")})
	}
	logger.Debug("Successfully updated resources")

	newVersion, err := common.GetCurrentVersion(params.ProjectName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not retrieve latest version")})
	}
	return stage_resource.NewPutProjectProjectNameStageStageNameResourceCreated().WithPayload(&models.Version{
		Version: newVersion,
	})
}

// PutProjectProjectNameStageStageNameResourceResourceURIHandlerFunc updates the specified stage resource
func PutProjectProjectNameStageStageNameResourceResourceURIHandlerFunc(params stage_resource.PutProjectProjectNameStageStageNameResourceResourceURIParams) middleware.Responder {
	common.LockProject(params.ProjectName)
	defer common.UnlockProject(params.ProjectName)
	logger := utils.NewLogger("", "", "configuration-service")
	if !common.StageExists(params.ProjectName, params.StageName) {
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceResourceURIBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Stage does not exist")})
	}
	projectConfigPath := config.ConfigDir + "/" + params.ProjectName

	logger.Debug("Creating new resource(s) in: " + projectConfigPath + " in stage " + params.StageName)
	logger.Debug("Checking out branch: " + params.StageName)
	err := common.CheckoutBranch(params.ProjectName, params.StageName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceResourceURIBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not check out branch")})
	}

	filePath := projectConfigPath + "/" + params.ResourceURI
	common.WriteBase64EncodedFile(filePath, params.Resource.ResourceContent)

	logger.Debug("Staging Changes")
	err = common.StageAndCommitAll(params.ProjectName, "Updated resource: "+params.ResourceURI)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceResourceURIBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not commit changes")})
	}
	logger.Debug("Successfully updated resource: " + params.ResourceURI)

	newVersion, err := common.GetCurrentVersion(params.ProjectName)
	if err != nil {
		logger.Error(err.Error())
		return stage_resource.NewPutProjectProjectNameStageStageNameResourceResourceURIBadRequest().WithPayload(&models.Error{Code: 400, Message: swag.String("Could not retrieve latest version")})
	}
	return stage_resource.NewPutProjectProjectNameStageStageNameResourceResourceURICreated().WithPayload(&models.Version{
		Version: newVersion,
	})
}

// DeleteProjectProjectNameStageStageNameResourceResourceURIHandlerFunc deletes the specified stage resource
func DeleteProjectProjectNameStageStageNameResourceResourceURIHandlerFunc(params stage_resource.DeleteProjectProjectNameStageStageNameResourceResourceURIParams) middleware.Responder {
	return middleware.NotImplemented("operation stage_resource.DeleteProjectProjectNameStageStageNameResourceResourceURI has not yet been implemented")
}
