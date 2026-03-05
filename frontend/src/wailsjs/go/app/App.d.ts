export interface ScreenNameItem {
  id: number;
  name: string;
}

export interface ConfigData {
  // 扫描配置
  scanInterval: number;
  dataDir: string;
  // 代理配置
  httpProxy: string;
  httpsProxy: string;
  noProxy: string;
  // 数据库配置
  dbHost: string;
  dbPort: number;
  dbUsername: string;
  dbPassword: string;
  dbDatabase: string;
  dbCharset: string;
}

export function GetAllScreenNames(): Promise<ScreenNameItem[]>;
export function AddScreenNames(input: string): Promise<number>;
export function DeleteScreenName(name: string): Promise<void>;
export function StartScan(): Promise<void>;
export function StopScan(): Promise<void>;
export function IsScanning(): Promise<boolean>;
export function RunOnce(): Promise<void>;
export function GetConfig(): Promise<ConfigData>;
export function SaveConfig(config: ConfigData): Promise<void>;
export function OpenConfigDir(): Promise<void>;
export function GetInitError(): Promise<string>;
