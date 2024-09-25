package fileman

import (
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *FileMan) sendAction(actionRequest model.HostAction) (*model.ActionResponse, error) {
	var actionResponse model.ActionResponse

	resp, err := s.restyClient.R().
		SetBody(actionRequest).
		SetResult(&actionResponse).
		Post("http://127.0.0.1:8080/actions") // 修改URL路径

	if err != nil {
		global.LOG.Error("failed to send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		global.LOG.Error("received error response: %s", resp.Status())
		return nil, fmt.Errorf("received error response: %s", resp.Status())
	}

	return &actionResponse, nil
}

func (s *FileMan) getFileTree(op model.FileOption) ([]model.FileTree, error) {
	var fileTree []model.FileTree

	data, err := utils.ToJSONString(op)
	if err != nil {
		return fileTree, err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Tree,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return fileTree, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fileTree, fmt.Errorf("failed to get filetree")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileTree)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to filetree: %v", err)
		return fileTree, fmt.Errorf("json err: %v", err)
	}

	return fileTree, nil
}

func (s *FileMan) getFileList(op model.FileOption) (*model.FileInfo, error) {
	var fileInfo model.FileInfo
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &fileInfo, err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_List,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &fileInfo, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &fileInfo, fmt.Errorf("failed to get file list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file list: %v", err)
		return &fileInfo, fmt.Errorf("json err: %v", err)
	}

	return &fileInfo, nil
}

func (s *FileMan) create(op model.FileCreate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to get file list")
	}

	return nil
}
func (s *FileMan) delete(op model.FileDelete) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to delete")
	}

	return nil
}
func (s *FileMan) batchDelete(op model.FileBatchDelete) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Batch_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to batch delete")
	}

	return nil
}
func (s *FileMan) compress(op model.FileCompress) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Compress,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to compress")
	}

	return nil
}
func (s *FileMan) decompress(op model.FileDeCompress) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Decompress,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to decompress")
	}

	return nil
}
func (s *FileMan) getContent(op model.FileContentReq) (*model.FileInfo, error) {
	var fileInfo model.FileInfo
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &fileInfo, err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &fileInfo, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &fileInfo, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &fileInfo, fmt.Errorf("json err: %v", err)
	}

	return &fileInfo, nil
}
func (s *FileMan) saveContent(op model.FileEdit) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Content_Modify,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to edit content")
	}

	return nil
}

func (s *FileMan) uploadFile(hostID uint, path string, file *multipart.FileHeader) error {
	return conn.CENTER.UploadFile(hostID, path, file)
}

func (s *FileMan) downloadFile(c *gin.Context, hostID uint, path string) error {
	return conn.CENTER.DownloadFile(c, hostID, path)
}

func (s *FileMan) dirSize(op model.DirSizeReq) (*model.DirSizeRes, error) {
	var dirSize model.DirSizeRes
	data, err := utils.ToJSONString(op)
	if err != nil {
		return &dirSize, err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Dir_Size,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &dirSize, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &dirSize, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &dirSize)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &dirSize, fmt.Errorf("json err: %v", err)
	}

	return &dirSize, nil
}
func (s *FileMan) changeName(op model.FileRename) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Change_Name,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to change name")
	}

	return nil
}
func (s *FileMan) mvFile(op model.FileMove) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Move,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to mv file")
	}

	return nil
}
func (s *FileMan) changeOwner(op model.FileRoleUpdate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Change_Owner,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to change owner")
	}

	return nil
}
func (s *FileMan) changeMode(op model.FileCreate) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Change_Mode,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to change mode")
	}

	return nil
}
func (s *FileMan) batchChangeModeAndOwner(op model.FileRoleReq) error {
	data, err := utils.ToJSONString(op)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: op.HostID,
		Action: model.Action{
			Action: model.File_Batch_Change_Owner,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to batch change owner")
	}

	return nil
}

func (s *FileMan) getFavoriteList(req model.FavoriteListReq) (*model.PageResult, error) {
	var pageResult model.PageResult
	data, err := utils.ToJSONString(req.PageInfo)
	if err != nil {
		return &pageResult, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Favorite_List,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &pageResult, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &pageResult, fmt.Errorf("failed to get fav list")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &pageResult)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to fav list: %v", err)
		return &pageResult, fmt.Errorf("json err: %v", err)
	}

	return &pageResult, nil
}
func (s *FileMan) createFavorite(req model.FavoriteCreate) (*model.Favorite, error) {
	var favorite model.Favorite
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &favorite, err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Favorite_Create,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &favorite, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &favorite, fmt.Errorf("failed to create fav")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &favorite)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file list: %v", err)
		return &favorite, fmt.Errorf("json err: %v", err)
	}

	return &favorite, nil
}

func (s *FileMan) deleteFavorite(req model.FavoriteDelete) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: req.HostID,
		Action: model.Action{
			Action: model.Favorite_Delete,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to delete fav")
	}

	return nil
}
