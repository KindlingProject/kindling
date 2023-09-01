import { useLocation } from 'react-router-dom';

// 获取url上带的params参数。xxx?A=a&B=b
export function getParamsFromSearchParams<T = any>(search: string): T {
    const searchParams = new URLSearchParams(search);
    const params = {};
    for (const [key, value] of searchParams.entries()) {
        params[key] = value;
    }
    return params as T;
}

export function useSearchParams<T = any>(): T {
    const { search } = useLocation()
    return getParamsFromSearchParams(search)
}