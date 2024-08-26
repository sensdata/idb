package model

const (
	SysInfo_OverView string = "sysinfo_overview"
	SysInfo_Network  string = "sysinfo_network"
	SysInfo_System   string = "sysinfo_system"

	File_List               string = "file_list"
	File_Create             string = "file_create"
	File_Delete             string = "file_delete"
	File_Batch_Delete       string = "file_batch_delete"
	File_Batch_Change_Owner string = "file_batch_change_owner"
	File_Change_Mode        string = "file_change_mode"
	File_Change_Owner       string = "file_change_owner"
	File_Change_Name        string = "file_change_name"
	File_Compress           string = "file_compress"
	File_Decompress         string = "file_decompress"
	File_Content            string = "file_content"
	File_Content_Modify     string = "file_content_modify"
	File_Move               string = "file_move"
	File_Dir_Size           string = "file_dir_size"
)

// Action消息结构
type Action struct {
	Action string `json:"action"`
	Result bool   `json:"result"`
	Data   string `json:"data"`
}

type HostAction struct {
	HostID uint   `json:"hostId"`
	Action Action `json:"action"`
}

type ActionResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    HostAction `json:"data"`
}
