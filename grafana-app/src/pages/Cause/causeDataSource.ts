import { getBackendSrv, getAppEvents } from '@grafana/runtime';
import { AppEvents } from '@grafana/data';

const appEvents = getAppEvents();
class CauseDataSource {
    constructor() {}
    esServer= 'http://10.10.103.96:8081';

    excuteQuery = async(api: string, params?: { [key: string]: string }) => {
        try {
            const response: any = await getBackendSrv().get(`${this.esServer}${api}${params?.length ? `?${params}` : ''}`);
            console.log('response', response);
            return Promise.resolve(response);
        } catch(err: any) {
            console.error('Error', err);
            appEvents.publish({
                type: AppEvents.alertWarning.name,
                payload: [err.data.errorMsg],
            });
            return Promise.reject(err);
        }
    }

    excutePostQuery = async(api: string, params?: { [key: string]: any }) => {
        try {
            const response: any = await getBackendSrv().post(`${this.esServer}${api}`, params);
            console.log('response', response);
            return Promise.resolve(response);
        } catch(err: any) {
            console.error('Error', err);
            appEvents.publish({
                type: AppEvents.alertWarning.name,
                payload: [err.data.errorMsg],
            });
            return Promise.reject(err);
        }
    }
}
export default CauseDataSource;
