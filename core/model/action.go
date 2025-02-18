package model

const (
	Host_Status string = "host_status"

	SysInfo_OverView      string = "sysinfo_overview"
	SysInfo_Network       string = "sysinfo_network"
	SysInfo_System        string = "sysinfo_system"
	SysInfo_Set_Time      string = "sysinfo_set_time"
	SysInfo_Set_Time_Zone string = "sysinfo_set_time_zone"
	SysInfo_Sync_Time     string = "sysinfo_sync_time"

	File_Tree               string = "file_tree"
	File_List               string = "file_list"
	File_Search             string = "file_search"
	File_Create             string = "file_create"
	File_Delete             string = "file_delete"
	File_Batch_Delete       string = "file_batch_delete"
	File_Batch_Change_Owner string = "file_batch_change_owner"
	File_Batch_Change_Mode  string = "file_batch_change_mode"
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

	Git_Init      string = "git_init"
	Git_File_List string = "git_file_list"
	Git_File      string = "git_file"
	Git_Create    string = "git_create"
	Git_Update    string = "git_update"
	Git_Delete    string = "git_delete"
	Git_Restore   string = "git_restore"
	Git_Log       string = "git_log"
	Git_Diff      string = "git_diff"

	Script_Exec string = "script_exec"

	Docker_Status                   string = "docker_status"
	Docker_Conf                     string = "docker_conf"
	Docker_Upd_Conf                 string = "docker_upd_conf"
	Docker_Upd_Conf_File            string = "docker_upd_conf_file"
	Docker_Upd_Log                  string = "docker_upd_log"
	Docker_Upd_Ipv6                 string = "docker_upd_ipv6"
	Docker_Operation                string = "docker_operation"
	Docker_Inspect                  string = "docker_inspect"
	Docker_Prune                    string = "docker_prune"
	Docker_Container_Query          string = "docker_container_query"
	Docker_Container_Names          string = "docker_container_names"
	Docker_Container_Create         string = "docker_container_create"
	Docker_Container_Update         string = "docker_container_update"
	Docker_Container_Upgrade        string = "docker_container_upgrade"
	Docker_Container_Info           string = "docker_container_info"
	Docker_Container_Resource_Usage string = "docker_container_resource_usage"
	Docker_Container_Resource_Limit string = "docker_container_resource_limit"
	Docker_Container_Stats          string = "docker_container_stats"
	Docker_Container_Rename         string = "docker_container_rename"
	Docker_Container_Log_Clean      string = "docker_container_log_clean"
	Docker_Container_Operation      string = "docker_container_operation"
	Docker_Container_Logs           string = "docker_container_logs"
	Docker_Image_Page               string = "docker_image_page"
	Docker_Image_List               string = "docker_image_list"
	Docker_Image_Build              string = "docker_image_build"
	Docker_Image_Pull               string = "docker_image_pull"
	Docker_Image_Load               string = "docker_image_load"
	Docker_Image_Save               string = "docker_image_save"
	Docker_Image_Push               string = "docker_image_push"
	Docker_Image_Remove             string = "docker_image_remove"
	Docker_Image_Tag                string = "docker_image_tag"
	Docker_Volume_Page              string = "docker_volume_page"
	Docker_Volume_List              string = "docker_volume_list"
	Docker_Volume_Delete            string = "docker_volume_delete"
	Docker_Volume_Create            string = "docker_volume_create"
	Docker_Network_Page             string = "docker_network_page"
	Docker_Network_List             string = "docker_network_list"
	Docker_Network_Delete           string = "docker_network_delete"
	Docker_Network_Create           string = "docker_network_create"
	Docker_Compose_Page             string = "docker_compose_page"
	Docker_Compose_Create           string = "docker_compose_create"
	Docker_Compose_Operation        string = "docker_compose_operation"
	Docker_Compose_Test             string = "docker_compose_test"
	Docker_Compose_Update           string = "docker_compose_update"

	CA_Groups       string = "ca_groups"
	CA_Group_Pk     string = "ca_group_pk"
	CA_Group_Csr    string = "ca_group_csr"
	CA_Group_Create string = "ca_group_create"
	CA_Group_Remove string = "ca_group_remove"
	CA_Self_Sign    string = "ca_self_sign"
	CA_Info         string = "ca_info"
	CA_Complete     string = "ca_complete"
	CA_Remove       string = "ca_remove"
	CA_Import       string = "ca_import"

	Terminal_List    string = "terminal_list"
	Terminal_Detach  string = "terminal_detach"
	Terminal_Finish  string = "terminal_finish"
	Terminal_Rename  string = "terminal_rename"
	Terminal_Install string = "terminal_install"
	Terminal_Prune   string = "terminal_prune"
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
