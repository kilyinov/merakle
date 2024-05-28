import { error } from '@sveltejs/kit';

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
    console.log('Loading portal page');
    const cells = await getCellsByUsername('KonstantinIlinov');
    return {
        title: 'Hello world!',
        content: 'Welcome to admin portal.'
    };

    // error(404, 'Not found');
}

async function getCellsByUsername({fetch:any, username:string}) {
    const res = await fetch(`http://localhost:8080/cells/my`);
	const item = await res.json();

	return { item };
}