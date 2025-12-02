export interface AppFormValidation {
  MinLength: number;
  MaxLength: number;
  Pattern: string;
  MinValue: number;
  MaxValue: number;
}

export interface AppFormField {
  Name: string;
  Label: string;
  Key: string;
  Type: string;
  Default: string;
  Required: boolean;
  Hint: string;
  Options: string[];
  Validation: AppFormValidation;
}

export interface AppEntity {
  id: number;
  type: string;
  name: string;
  display_name: string;
  current_version: string;
  category: string;
  tags: string[];
  title: string;
  description: string;
  vendor: {
    name: string;
    url: string;
  };
  packager: {
    name: string;
    url: string;
  };
  has_upgrade: boolean;
  status: string;
  versions: Array<{
    id: number;
    version: string;
    update_version: string;
    compose_content: string;
    env_content: string;
    status: string;
    created_at: string;
    can_upgrade?: boolean;
  }>;
  form: {
    Fields: AppFormField[];
  };
}

export type AppSimpleEntity = Omit<AppEntity, 'form'>;
