import { redirect } from '@sveltejs/kit';
import { checkAuth } from '$lib/stores/auth';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ url, cookies }) => {
    const user = await checkAuth(cookies);

    const path = url.pathname.replace(/\/+$/, '');
    const isPublic = path === '/login' || path.startsWith('/sessions');

    if (!user && !isPublic) {
        throw redirect(303, '/login');
    }

    return {
        user
    };
};
