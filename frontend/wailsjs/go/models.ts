export namespace main {
	
	export class DNSProfile {
	    id: string;
	    name: string;
	    nameFa: string;
	    ipv4: string[];
	    ipv6: string[];
	    isPreset: boolean;
	    isAutomatic: boolean;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new DNSProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.nameFa = source["nameFa"];
	        this.ipv4 = source["ipv4"];
	        this.ipv6 = source["ipv6"];
	        this.isPreset = source["isPreset"];
	        this.isAutomatic = source["isAutomatic"];
	        this.color = source["color"];
	    }
	}
	export class AppSettings {
	    language: string;
	    theme: string;
	    favorites: string[];
	    customProfiles: DNSProfile[];
	    lastInterface: string;
	    applyToAll: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.favorites = source["favorites"];
	        this.customProfiles = this.convertValues(source["customProfiles"], DNSProfile);
	        this.lastInterface = source["lastInterface"];
	        this.applyToAll = source["applyToAll"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ApplyResult {
	    success: boolean;
	    code: string;
	    message: string;
	    needsElevation: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApplyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.needsElevation = source["needsElevation"];
	    }
	}
	
	export class NetworkInterface {
	    name: string;
	    displayName: string;
	    isUp: boolean;
	    mtu: number;
	    ipv4: string[];
	    ipv6: string[];
	    dns: string[];
	    dhcp: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NetworkInterface(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.isUp = source["isUp"];
	        this.mtu = source["mtu"];
	        this.ipv4 = source["ipv4"];
	        this.ipv6 = source["ipv6"];
	        this.dns = source["dns"];
	        this.dhcp = source["dhcp"];
	    }
	}
	export class PingResult {
	    profileId: string;
	    server: string;
	    latencyMs: number;
	    success: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new PingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileId = source["profileId"];
	        this.server = source["server"];
	        this.latencyMs = source["latencyMs"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}

}

