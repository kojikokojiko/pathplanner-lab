// Auth0 PKCE helpers

const AUTH0_DOMAIN = import.meta.env.VITE_AUTH0_DOMAIN ?? '';
const AUTH0_CLIENT_ID = import.meta.env.VITE_AUTH0_CLIENT_ID ?? '';
const APP_URL = import.meta.env.VITE_APP_URL ?? window.location.origin;
const REDIRECT_URI = `${APP_URL}/auth/callback`;

export const isAuth0Enabled = Boolean(AUTH0_DOMAIN && AUTH0_CLIENT_ID);

function base64UrlEncode(buffer: ArrayBuffer): string {
  return btoa(String.fromCharCode(...new Uint8Array(buffer)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

function generateRandom(length: number): string {
  const array = new Uint8Array(length);
  crypto.getRandomValues(array);
  return base64UrlEncode(array.buffer);
}

async function sha256(plain: string): Promise<ArrayBuffer> {
  const encoder = new TextEncoder();
  return crypto.subtle.digest('SHA-256', encoder.encode(plain));
}

export async function startLoginFlow(): Promise<void> {
  const verifier = generateRandom(64);
  const challenge = base64UrlEncode(await sha256(verifier));
  const state = generateRandom(16);

  sessionStorage.setItem('pkce_verifier', verifier);
  sessionStorage.setItem('pkce_state', state);

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: AUTH0_CLIENT_ID,
    redirect_uri: REDIRECT_URI,
    scope: 'openid email profile',
    code_challenge: challenge,
    code_challenge_method: 'S256',
    state,
  });

  window.location.href = `https://${AUTH0_DOMAIN}/authorize?${params}`;
}

export async function handleCallback(code: string, returnedState: string): Promise<string> {
  const verifier = sessionStorage.getItem('pkce_verifier');
  const expectedState = sessionStorage.getItem('pkce_state');

  if (!verifier || returnedState !== expectedState) {
    throw new Error('Invalid state or missing PKCE verifier');
  }

  sessionStorage.removeItem('pkce_verifier');
  sessionStorage.removeItem('pkce_state');

  const body = new URLSearchParams({
    grant_type: 'authorization_code',
    client_id: AUTH0_CLIENT_ID,
    code,
    redirect_uri: REDIRECT_URI,
    code_verifier: verifier,
  });

  const res = await fetch(`https://${AUTH0_DOMAIN}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: body.toString(),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Token exchange failed: ${text}`);
  }

  const data = await res.json();
  return data.id_token as string;
}
