export namespace app {
	
	export class ConfigData {
	    scanInterval: number;
	    dataDir: string;
	    httpProxy: string;
	    httpsProxy: string;
	    noProxy: string;
	    dbHost: string;
	    dbPort: number;
	    dbUsername: string;
	    dbPassword: string;
	    dbDatabase: string;
	    dbCharset: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scanInterval = source["scanInterval"];
	        this.dataDir = source["dataDir"];
	        this.httpProxy = source["httpProxy"];
	        this.httpsProxy = source["httpsProxy"];
	        this.noProxy = source["noProxy"];
	        this.dbHost = source["dbHost"];
	        this.dbPort = source["dbPort"];
	        this.dbUsername = source["dbUsername"];
	        this.dbPassword = source["dbPassword"];
	        this.dbDatabase = source["dbDatabase"];
	        this.dbCharset = source["dbCharset"];
	    }
	}
	export class ScreenNameItem {
	    id: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ScreenNameItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

