const firebaseLogin = async (email, password) => {
  const apiKey = process.env.REACT_APP_FIREBASE_API_KEY;
  const url = `https://www.googleapis.com/identitytoolkit/v3/relyingparty/verifyPassword?key=${apiKey}`;

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email: email,
        password: password,
        returnSecureToken: true,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      const errorMessage = errorData.error.message || "Login failed";
      throw new Error(errorMessage);
    }

    const data = await response.json();

    const user = {
      email: data.email,
      displayName: data.displayName,
      photoUrl: data.photoUrl,
      uid: data.localId,
      expiresIn: data.expiresIn,
      storedAt: Date.now(),
    };

    return { user, idToken: data.idToken };
  } catch (error) {
    throw error;
  }
};

export default firebaseLogin;
