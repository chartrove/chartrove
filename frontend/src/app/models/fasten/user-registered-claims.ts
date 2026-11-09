export class UserRegisteredClaims {
  iss: string //Issuer
  sub: string //Subject
  aud?: string //Audience
  exp: number //ExpiresAt (unix timestamp)
  nbf?: number //NotBefore (unix timestamp)
  iat?: number //IssuedAt (unix timestamp)
  id?: string //ID
  native_user_id?: string // Immutable local identity; server verification remains authoritative.

  full_name: string //FullName
  picture: string //Picture
  email: string //Email
  role: string //Role
}
