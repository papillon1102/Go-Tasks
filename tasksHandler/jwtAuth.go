package handler

// JWT Middleware (FIXME)
// func (ah *AuthHandler) AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		// Get token value from header
// 		tokenValue := c.GetHeader("Authorization")
// 		claims := &Claims{}

// 		// Parse token value
// 		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
// 			return []byte(os.Getenv("JWT_SECRET")), nil
// 		})

// 		if err != nil {
// 			log.Debug().Msgf("err", err)
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}

// 		// if token not valid => return unauthor status
// 		if !tkn.Valid {
// 			log.Debug().Msg("token validerr")
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 		}

// 		c.Next()
// 	}
// }

// JWT Signin-Handler (FIXME)
// func (au *AuthHandler) SignInHandler(c *gin.Context) {
// 	var authUser models.AuthUser
// 	if err := c.ShouldBindJSON(&authUser); err != nil {
// 		log.Error().Err(err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if authUser.Username != "admin" || authUser.Password != "password" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
// 		return
// 	}

// 	// Create expiration time & claims
// 	expirationTime := time.Now().Add(time.Minute * 10)
// 	claims := &Claims{
// 		Username: authUser.Username,
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	// Create new token
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError,
// 			gin.H{"error": err.Error()})
// 		return
// 	}
// 	jwtOutput := PWTOutput{
// 		Token:   tokenString,
// 		Expires: expirationTime,
// 	}
// 	c.JSON(http.StatusOK, jwtOutput)
// }
