package model

const (
	SysInfo_OverView string = "sysinfo_overview"
	SysInfo_Network  string = "sysinfo_network"
	SysInfo_System   string = "sysinfo_system"

	File_Tree               string = "file_tree"
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
	File_Upload             string = "file_upload"
	File_Download           string = "file_download"
	Favorite_List           string = "favorite_list"
	Favorite_Create         string = "favorite_create"
	Favorite_Delete         string = "favorite_delete"

	Ssh_Config                string = "ssh_config"
	Ssh_Config_Update         string = "ssh_config_update"
	Ssh_Config_Content        string = "ssh_config_content"
	Ssh_Config_Content_Update string = "ssh_config_content_update"
	Ssh_Operate               string = "ssh_operate"
	Ssh_Secret                string = "ssh_secret"
	Ssh_Secret_Create         string = "ssh_secret_create"
	Ssh_Log                   string = "ssh_log"

	Script_Exec string = "script_exec"

	Git_Init      string = "git_init"
	Git_File_List string = "git_file_list"
	Git_File      string = "git_file"
	Git_Create    string = "git_create"
	Git_Update    string = "git_update"
	Git_Delete    string = "git_delete"
	Git_Restore   string = "git_restore"
	Git_Log       string = "git_log"
	Git_Diff      string = "git_diff"
)

// Action消息结构
type Action struct {
	Action string `json:"action"`
	Result bool   `json:"result"`
	Data   string `json:"data"`
}

type HostAction struct {
	HostID uint   `json:"host_id"`
	Action Action `json:"action"`
}

type ActionResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    HostAction `json:"data"`
}
