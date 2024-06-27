#!/bin/bash

# Define user details
USERNAME="theaddicts"
USERID="1001"  # Adjust accordingly to match your actual user ID
GROUPID="1001" # Adjust accordingly to match your actual group ID
FULLNAME="The Addicts"
HOMEDIR="/home/theaddicts"
SHELL="/bin/bash"

# Define paths to custom etc directory
CUSTOM_ETC_DIR="/home/theaddicts/custom_etc"
CUSTOM_PASSWD="$CUSTOM_ETC_DIR/passwd"
CUSTOM_GROUP="$CUSTOM_ETC_DIR/group"
CHROOT_DIR="/home/theaddicts/my_chroot"

# Ensure the custom etc directory and chroot directory exist
mkdir -p $CUSTOM_ETC_DIR
mkdir -p $CHROOT_DIR

# Create custom passwd and group files if they don't exist
if [ ! -f $CUSTOM_PASSWD ]; then
  cp /etc/passwd $CUSTOM_PASSWD
fi

if [ ! -f $CUSTOM_GROUP ]; then
  cp /etc/group $CUSTOM_GROUP
fi

# Function to add or update user in /etc/passwd
add_or_update_user() {
  if grep -q "^$USERNAME:" $CUSTOM_PASSWD; then
    # Update existing user entry
    sed -i "s/^$USERNAME:.*/$USERNAME:x:$USERID:$GROUPID:$FULLNAME:$HOMEDIR:$SHELL/" $CUSTOM_PASSWD
  else
    # Add new user entry
    echo "$USERNAME:x:$USERID:$GROUPID:$FULLNAME:$HOMEDIR:$SHELL" >> $CUSTOM_PASSWD
  fi
}

# Function to add or update group in /etc/group
add_or_update_group() {
  if grep -q "^$USERNAME:" $CUSTOM_GROUP; then
    # Update existing group entry
    sed -i "s/^$USERNAME:.*/$USERNAME:x:$GROUPID:/" $CUSTOM_GROUP
  else
    # Add new group entry
    echo "$USERNAME:x:$GROUPID:" >> $CUSTOM_GROUP
  fi
}

# Add or update user and group
add_or_update_user
add_or_update_group

# Create symlinks for the custom etc files
ln -sf $CUSTOM_PASSWD $CHROOT_DIR/etc/passwd
ln -sf $CUSTOM_GROUP $CHROOT_DIR/etc/group

# Create symlinks for necessary directories and binaries
for dir in bin lib lib64 usr etc var run tmp; do
  mkdir -p $CHROOT_DIR/$dir
  ln -sf /$dir $CHROOT_DIR/$dir
done

# Create necessary directories inside the chroot
mkdir -p $CHROOT_DIR/dev
mkdir -p $CHROOT_DIR/proc
mkdir -p $CHROOT_DIR/sys
mkdir -p $CHROOT_DIR/home
mkdir -p $CHROOT_DIR/tmp

# Bind mount necessary directories (using user permissions, assuming `bindfs` is installed)
bindfs --map=root/$(id -u) /dev $CHROOT_DIR/dev
bindfs --map=root/$(id -u) /proc $CHROOT_DIR/proc
bindfs --map=root/$(id -u) /sys $CHROOT_DIR/sys
bindfs --map=root/$(id -u) /tmp $CHROOT_DIR/tmp
bindfs --map=root/$(id -u) /run $CHROOT_DIR/run

# Run your application in the chroot environment
chroot $CHROOT_DIR /bin/bash -c "your_application_command_here"

# Note: You need to install `bindfs` if it's not already installed
# You can install it using your package manager, for example:
# sudo apt-get install bindfs (if you have permissions)
