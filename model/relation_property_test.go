package model

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zhoudm1743/torm/db"
)

// **Property 6: Relationship Data Consistency**
// **Validates: Requirements 6.1-6.8**
//
// *For any* Model with defined relationships (HasOne, HasMany, BelongsTo, BelongsToMany),
// eager loading via With() SHALL return the same related data as lazy loading via
// relationship accessor.

// setupRelationTestTables creates test tables for relationship testing
func setupRelationTestTables() error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}

	// Users table
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Profiles table (HasOne relationship)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		bio TEXT,
		avatar TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create profiles table: %w", err)
	}

	// Posts table (HasMany relationship)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	// Comments table (BelongsTo relationship)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	// Roles table (BelongsToMany relationship)
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	if err != nil {
		return fmt.Errorf("failed to create roles table: %w", err)
	}

	// User-Role pivot table
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS role_user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		role_id INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (role_id) REFERENCES roles(id)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create role_user table: %w", err)
	}

	return nil
}

// cleanupRelationTestTables removes all data from test tables
func cleanupRelationTestTables() error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}

	tables := []string{"role_user", "comments", "posts", "profiles", "roles", "users"}
	for _, table := range tables {
		_, err = conn.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			return fmt.Errorf("failed to clean %s: %w", table, err)
		}
	}
	return nil
}

// dropRelationTestTables drops all test tables
func dropRelationTestTables() error {
	conn, err := db.DB("default")
	if err != nil {
		return err
	}

	tables := []string{"role_user", "comments", "posts", "profiles", "roles", "users"}
	for _, table := range tables {
		_, err = conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return fmt.Errorf("failed to drop %s: %w", table, err)
		}
	}
	return nil
}

// createTestUser creates a test user and returns the ID
func createTestUser(name, email string) (int64, error) {
	user := NewModel("users")
	user.SetAttribute("name", name)
	user.SetAttribute("email", email)
	err := user.Save()
	if err != nil {
		return 0, err
	}
	id, ok := user.GetKey().(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected key type: %T", user.GetKey())
	}
	return id, nil
}

// createTestProfile creates a test profile for a user
func createTestProfile(userID int64, bio, avatar string) (int64, error) {
	profile := NewModel("profiles")
	profile.SetAttribute("user_id", userID)
	profile.SetAttribute("bio", bio)
	profile.SetAttribute("avatar", avatar)
	err := profile.Save()
	if err != nil {
		return 0, err
	}
	id, ok := profile.GetKey().(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected key type: %T", profile.GetKey())
	}
	return id, nil
}

// createTestPost creates a test post for a user
func createTestPost(userID int64, title, content string) (int64, error) {
	post := NewModel("posts")
	post.SetAttribute("user_id", userID)
	post.SetAttribute("title", title)
	post.SetAttribute("content", content)
	err := post.Save()
	if err != nil {
		return 0, err
	}
	id, ok := post.GetKey().(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected key type: %T", post.GetKey())
	}
	return id, nil
}

// createTestComment creates a test comment
func createTestComment(postID, userID int64, content string) (int64, error) {
	comment := NewModel("comments")
	comment.SetAttribute("post_id", postID)
	comment.SetAttribute("user_id", userID)
	comment.SetAttribute("content", content)
	err := comment.Save()
	if err != nil {
		return 0, err
	}
	id, ok := comment.GetKey().(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected key type: %T", comment.GetKey())
	}
	return id, nil
}

// createTestRole creates a test role
func createTestRole(name, description string) (int64, error) {
	role := NewModel("roles")
	role.SetAttribute("name", name)
	role.SetAttribute("description", description)
	err := role.Save()
	if err != nil {
		return 0, err
	}
	id, ok := role.GetKey().(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected key type: %T", role.GetKey())
	}
	return id, nil
}

// attachRoleToUser attaches a role to a user
func attachRoleToUser(userID, roleID int64) error {
	pivot := NewModel("role_user")
	pivot.DisableTimestamps() // Pivot table doesn't have timestamp columns
	pivot.SetAttribute("user_id", userID)
	pivot.SetAttribute("role_id", roleID)
	return pivot.Save()
}

// TestProperty6_HasOneRelationship tests HasOne relationship consistency
// Validates: Requirements 6.1
func TestProperty6_HasOneRelationship(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	err := setupRelationTestTables()
	if err != nil {
		t.Fatalf("Failed to setup test tables: %v", err)
	}
	defer dropRelationTestTables()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 6.1: HasOne returns single related record
	properties.Property("HasOne returns single related record", prop.ForAll(
		func(userName string, bio string) bool {
			cleanupRelationTestTables()

			// Create user
			userID, err := createTestUser(userName, userName+"@test.com")
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create profile for user
			_, err = createTestProfile(userID, bio, "avatar.png")
			if err != nil {
				t.Logf("Failed to create profile: %v", err)
				return false
			}

			// Load user and get profile via HasOne
			user := NewModel("users")
			err = user.Find(userID)
			if err != nil {
				t.Logf("Failed to find user: %v", err)
				return false
			}

			// Get profile using HasOne relation
			profileRelation := user.HasOne(&BaseModel{}, "user_id", "id")
			profileRelation.relatedTable = "profiles"
			profileRelation.query, _ = db.NewQueryBuilder("default")
			profileRelation.query = profileRelation.query.From("profiles")

			profile, err := profileRelation.First()
			if err != nil {
				t.Logf("Failed to get profile: %v", err)
				return false
			}

			// Verify profile belongs to user
			if profile == nil {
				t.Log("Expected profile to be found")
				return false
			}

			profileUserID, ok := profile["user_id"].(int64)
			if !ok {
				t.Logf("user_id type mismatch: %T", profile["user_id"])
				return false
			}

			if profileUserID != userID {
				t.Logf("Profile user_id mismatch: expected %d, got %d", userID, profileUserID)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 100 }),
	))

	properties.TestingRun(t)
}

// TestProperty6_HasManyRelationship tests HasMany relationship consistency
// Validates: Requirements 6.2
func TestProperty6_HasManyRelationship(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	err := setupRelationTestTables()
	if err != nil {
		t.Fatalf("Failed to setup test tables: %v", err)
	}
	defer dropRelationTestTables()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 6.2: HasMany returns all related records
	properties.Property("HasMany returns all related records", prop.ForAll(
		func(userName string, postCount int) bool {
			cleanupRelationTestTables()

			// Create user
			userID, err := createTestUser(userName, userName+"@test.com")
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create multiple posts for user
			for i := 0; i < postCount; i++ {
				_, err = createTestPost(userID, fmt.Sprintf("Post %d", i), fmt.Sprintf("Content %d", i))
				if err != nil {
					t.Logf("Failed to create post: %v", err)
					return false
				}
			}

			// Load user and get posts via HasMany
			user := NewModel("users")
			err = user.Find(userID)
			if err != nil {
				t.Logf("Failed to find user: %v", err)
				return false
			}

			// Get posts using HasMany relation
			postsRelation := user.HasMany(&BaseModel{}, "user_id", "id")
			postsRelation.relatedTable = "posts"
			postsRelation.query, _ = db.NewQueryBuilder("default")
			postsRelation.query = postsRelation.query.From("posts")

			posts, err := postsRelation.Get()
			if err != nil {
				t.Logf("Failed to get posts: %v", err)
				return false
			}

			// Verify post count matches
			if len(posts) != postCount {
				t.Logf("Post count mismatch: expected %d, got %d", postCount, len(posts))
				return false
			}

			// Verify all posts belong to user
			for _, post := range posts {
				postUserID, ok := post["user_id"].(int64)
				if !ok {
					t.Logf("user_id type mismatch: %T", post["user_id"])
					return false
				}
				if postUserID != userID {
					t.Logf("Post user_id mismatch: expected %d, got %d", userID, postUserID)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}

// TestProperty6_BelongsToRelationship tests BelongsTo relationship consistency
// Validates: Requirements 6.3
func TestProperty6_BelongsToRelationship(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	err := setupRelationTestTables()
	if err != nil {
		t.Fatalf("Failed to setup test tables: %v", err)
	}
	defer dropRelationTestTables()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 6.3: BelongsTo returns parent record
	properties.Property("BelongsTo returns parent record", prop.ForAll(
		func(userName string, postTitle string) bool {
			cleanupRelationTestTables()

			// Create user
			userID, err := createTestUser(userName, userName+"@test.com")
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create post for user
			postID, err := createTestPost(userID, postTitle, "Content")
			if err != nil {
				t.Logf("Failed to create post: %v", err)
				return false
			}

			// Load post and get user via BelongsTo
			post := NewModel("posts")
			err = post.Find(postID)
			if err != nil {
				t.Logf("Failed to find post: %v", err)
				return false
			}

			// Get user using BelongsTo relation
			userRelation := post.BelongsTo(&BaseModel{}, "user_id", "id")
			userRelation.relatedTable = "users"
			userRelation.query, _ = db.NewQueryBuilder("default")
			userRelation.query = userRelation.query.From("users")

			user, err := userRelation.First()
			if err != nil {
				t.Logf("Failed to get user: %v", err)
				return false
			}

			// Verify user is found
			if user == nil {
				t.Log("Expected user to be found")
				return false
			}

			// Verify user ID matches
			foundUserID, ok := user["id"].(int64)
			if !ok {
				t.Logf("id type mismatch: %T", user["id"])
				return false
			}

			if foundUserID != userID {
				t.Logf("User ID mismatch: expected %d, got %d", userID, foundUserID)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t)
}

// TestProperty6_BelongsToManyRelationship tests BelongsToMany relationship consistency
// Validates: Requirements 6.4
func TestProperty6_BelongsToManyRelationship(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	err := setupRelationTestTables()
	if err != nil {
		t.Fatalf("Failed to setup test tables: %v", err)
	}
	defer dropRelationTestTables()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 6.4: BelongsToMany returns all related records via pivot table
	properties.Property("BelongsToMany returns all related records via pivot table", prop.ForAll(
		func(userName string, roleCount int) bool {
			cleanupRelationTestTables()

			// Create user
			userID, err := createTestUser(userName, userName+"@test.com")
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create roles and attach to user
			roleIDs := make([]int64, roleCount)
			for i := 0; i < roleCount; i++ {
				roleID, err := createTestRole(fmt.Sprintf("Role %d", i), fmt.Sprintf("Description %d", i))
				if err != nil {
					t.Logf("Failed to create role: %v", err)
					return false
				}
				roleIDs[i] = roleID

				err = attachRoleToUser(userID, roleID)
				if err != nil {
					t.Logf("Failed to attach role: %v", err)
					return false
				}
			}

			// Load user and get roles via BelongsToMany
			user := NewModel("users")
			err = user.Find(userID)
			if err != nil {
				t.Logf("Failed to find user: %v", err)
				return false
			}

			// Get roles using BelongsToMany relation
			rolesRelation := user.BelongsToMany(&BaseModel{}, "role_user", "role_id", "user_id")
			rolesRelation.relatedTable = "roles"
			rolesRelation.query, _ = db.NewQueryBuilder("default")
			rolesRelation.query = rolesRelation.query.From("roles")

			roles, err := rolesRelation.Get()
			if err != nil {
				t.Logf("Failed to get roles: %v", err)
				return false
			}

			// Verify role count matches
			if len(roles) != roleCount {
				t.Logf("Role count mismatch: expected %d, got %d", roleCount, len(roles))
				return false
			}

			// Verify all expected roles are present
			foundRoleIDs := make(map[int64]bool)
			for _, role := range roles {
				roleID, ok := role["id"].(int64)
				if !ok {
					t.Logf("id type mismatch: %T", role["id"])
					return false
				}
				foundRoleIDs[roleID] = true
			}

			for _, expectedID := range roleIDs {
				if !foundRoleIDs[expectedID] {
					t.Logf("Expected role ID %d not found", expectedID)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}

// TestProperty6_EagerLoadingConsistency tests eager loading vs lazy loading consistency
// Validates: Requirements 6.5, 6.6, 6.7
func TestProperty6_EagerLoadingConsistency(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	err := setupRelationTestTables()
	if err != nil {
		t.Fatalf("Failed to setup test tables: %v", err)
	}
	defer dropRelationTestTables()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 6.5: Eager loading returns same data as lazy loading for HasMany
	properties.Property("Eager loading returns same data as lazy loading for HasMany", prop.ForAll(
		func(userName string, postCount int) bool {
			cleanupRelationTestTables()

			// Create user
			userID, err := createTestUser(userName, userName+"@test.com")
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create posts
			postIDs := make([]int64, postCount)
			for i := 0; i < postCount; i++ {
				postID, err := createTestPost(userID, fmt.Sprintf("Post %d", i), fmt.Sprintf("Content %d", i))
				if err != nil {
					t.Logf("Failed to create post: %v", err)
					return false
				}
				postIDs[i] = postID
			}

			// Register the posts relation for eager loading
			RegisterRelation("users", "posts", RelationHasMany, func(parent *BaseModel) RelationInterface {
				relation := parent.HasMany(&BaseModel{}, "user_id", "id")
				relation.relatedTable = "posts"
				relation.query, _ = db.NewQueryBuilder("default")
				relation.query = relation.query.From("posts")
				return relation
			})

			// Get posts via lazy loading
			user := NewModel("users")
			err = user.Find(userID)
			if err != nil {
				t.Logf("Failed to find user: %v", err)
				return false
			}

			postsRelation := user.HasMany(&BaseModel{}, "user_id", "id")
			postsRelation.relatedTable = "posts"
			postsRelation.query, _ = db.NewQueryBuilder("default")
			postsRelation.query = postsRelation.query.From("posts")

			lazyPosts, err := postsRelation.Get()
			if err != nil {
				t.Logf("Failed to get posts via lazy loading: %v", err)
				return false
			}

			// Get posts via eager loading
			userData := []map[string]interface{}{user.GetAttributes()}
			eagerResults, err := postsRelation.EagerLoad(userData, "posts")
			if err != nil {
				t.Logf("Failed to eager load posts: %v", err)
				return false
			}

			// Compare results
			if len(eagerResults) != 1 {
				t.Logf("Expected 1 user result, got %d", len(eagerResults))
				return false
			}

			eagerPosts, ok := eagerResults[0]["posts"].([]map[string]interface{})
			if !ok {
				t.Logf("posts type mismatch: %T", eagerResults[0]["posts"])
				return false
			}

			if len(eagerPosts) != len(lazyPosts) {
				t.Logf("Post count mismatch: eager=%d, lazy=%d", len(eagerPosts), len(lazyPosts))
				return false
			}

			// Verify all post IDs match
			lazyPostIDs := make(map[int64]bool)
			for _, post := range lazyPosts {
				if id, ok := post["id"].(int64); ok {
					lazyPostIDs[id] = true
				}
			}

			for _, post := range eagerPosts {
				if id, ok := post["id"].(int64); ok {
					if !lazyPostIDs[id] {
						t.Logf("Eager loaded post ID %d not found in lazy loaded posts", id)
						return false
					}
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.IntRange(1, 5),
	))

	properties.TestingRun(t)
}

// TestProperty6_RelationshipConstraints tests relationship constraints
// Validates: Requirements 6.8
func TestProperty6_RelationshipConstraints(t *testing.T) {
	cleanup := setupDefaultManagerForTest(t)
	defer cleanup()

	err := setupRelationTestTables()
	if err != nil {
		t.Fatalf("Failed to setup test tables: %v", err)
	}
	defer dropRelationTestTables()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 6.6: Relationship constraints filter results correctly
	properties.Property("Relationship constraints filter results correctly", prop.ForAll(
		func(userName string) bool {
			cleanupRelationTestTables()

			// Create user
			userID, err := createTestUser(userName, userName+"@test.com")
			if err != nil {
				t.Logf("Failed to create user: %v", err)
				return false
			}

			// Create posts with different titles
			_, err = createTestPost(userID, "Important Post", "Content 1")
			if err != nil {
				t.Logf("Failed to create post 1: %v", err)
				return false
			}

			_, err = createTestPost(userID, "Regular Post", "Content 2")
			if err != nil {
				t.Logf("Failed to create post 2: %v", err)
				return false
			}

			_, err = createTestPost(userID, "Important Update", "Content 3")
			if err != nil {
				t.Logf("Failed to create post 3: %v", err)
				return false
			}

			// Load user
			user := NewModel("users")
			err = user.Find(userID)
			if err != nil {
				t.Logf("Failed to find user: %v", err)
				return false
			}

			// Get posts with constraint (title LIKE 'Important%')
			postsRelation := user.HasMany(&BaseModel{}, "user_id", "id")
			postsRelation.relatedTable = "posts"
			postsRelation.query, _ = db.NewQueryBuilder("default")
			postsRelation.query = postsRelation.query.From("posts")

			// Apply constraint
			postsRelation.Where("title", "LIKE", "Important%")

			posts, err := postsRelation.Get()
			if err != nil {
				t.Logf("Failed to get posts with constraint: %v", err)
				return false
			}

			// Should return only 2 posts (Important Post and Important Update)
			if len(posts) != 2 {
				t.Logf("Expected 2 posts with constraint, got %d", len(posts))
				return false
			}

			// Verify all returned posts have "Important" in title
			for _, post := range posts {
				title, ok := post["title"].(string)
				if !ok {
					t.Logf("title type mismatch: %T", post["title"])
					return false
				}
				if title != "Important Post" && title != "Important Update" {
					t.Logf("Unexpected post title: %s", title)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	properties.TestingRun(t)
}
