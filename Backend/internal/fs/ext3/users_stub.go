package ext3

import (
	"context"
	"fmt"

	"MIA_2S2025_P2_201905884/internal/fs"
)

// Stub implementations for P1 user/group management
// TODO: Implement full functionality by reading/writing to /users.txt

func (f *FS3) AddGroup(ctx context.Context, h fs.MountHandle, name string) error {
	// TODO: Implement
	// 1. Read /users.txt
	// 2. Check if group already exists
	// 3. Find next group ID
	// 4. Append line: <id>,G,<name>
	// 5. Write back to /users.txt
	// 6. Register operation in journal
	return fmt.Errorf("AddGroup not yet implemented in EXT3")
}

func (f *FS3) RemoveGroup(ctx context.Context, h fs.MountHandle, name string) error {
	// TODO: Implement
	// 1. Read /users.txt
	// 2. Find group line
	// 3. Mark as deleted (set status to 0)
	// 4. Write back to /users.txt
	// 5. Register operation in journal
	return fmt.Errorf("RemoveGroup not yet implemented in EXT3")
}

func (f *FS3) AddUser(ctx context.Context, h fs.MountHandle, user, pass, group string) error {
	// TODO: Implement
	// 1. Read /users.txt
	// 2. Check if user already exists
	// 3. Check if group exists
	// 4. Find next user ID
	// 5. Append line: <gid>,U,<user>,<group>,<pass>
	// 6. Write back to /users.txt
	// 7. Register operation in journal
	return fmt.Errorf("AddUser not yet implemented in EXT3")
}

func (f *FS3) RemoveUser(ctx context.Context, h fs.MountHandle, user string) error {
	// TODO: Implement
	// 1. Read /users.txt
	// 2. Find user line
	// 3. Mark as deleted (set status to 0)
	// 4. Write back to /users.txt
	// 5. Register operation in journal
	return fmt.Errorf("RemoveUser not yet implemented in EXT3")
}

func (f *FS3) ChangeUserGroup(ctx context.Context, h fs.MountHandle, user, group string) error {
	// TODO: Implement
	// 1. Read /users.txt
	// 2. Find user line
	// 3. Check if new group exists
	// 4. Update group field
	// 5. Write back to /users.txt
	// 6. Register operation in journal
	return fmt.Errorf("ChangeUserGroup not yet implemented in EXT3")
}
