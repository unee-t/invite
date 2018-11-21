# For any question about this script, ask Franck
#
#################################################################
#																#
# UPDATE THE BELOW VARIABLES ACCORDING TO YOUR NEEDS			#
#																#
#################################################################
#
# The MEFE invitation id that we want to process:
	SET @mefe_invitation_id = '%s';
#
# Environment: Which environment are you creating the unit in?
#	- 1 is for the DEV/Staging
#	- 2 is for the prod environment
#	- 3 is for the Demo environment
	SET @environment = %d;
#
########################################################################
#
#	ALL THE VARIABLES WE NEED HAVE BEEN DEFINED, WE CAN RUN THE SCRIPT #
#
########################################################################
#
#
#
#############################################
#											#
# IMPORTANT INFORMATION ABOUT THIS SCRIPT	#
#											#
#############################################
#
# Built for BZFE database v3.26+
#
# Use this script only if the Unit EXIST in the BZFE 
# It assumes that the unit has been created with all the necessary BZ objects and all the roles assigned to dummy users.
#
# Pre-requisite:
#	- The table 'ut_invitation_api_data' has been updated 
# 	- We know the MEFE Invitation id that we need to process.
#	- We know the environment where this script is run
# 
# This script depends on several SQL procedures:
#	- Procedure to add a user to a role in a unit.

# This script will also
#	- Add an existing BZ user as ASSIGNEE to an existing case which has already been created.
#	- Does NOT update the bug_user_last_visit table as the user had no action in there.
#
# Limits of this script:
#	- Unit must have all roles created with Dummy user roles.

####################
#
# Let's do this!
#
####################

	CALL `add_user_to_role_in_unit`;

# We have invited the user to a role in the unit
# We now invite the user to the case

	# Info about this script
		SET @this_script = '1_invite_user_to_a_case.sql';
	
	# Timestamp	
        SET @timestamp = NOW();

	# The case id
		SET @bz_case_id = (SELECT `bz_case_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

	# Do we need to change the case assignee?
		SET @change_case_assignee = IF (@invitation_type = 'type_assigned'
			, 1
			, 0
			)
			;

	# Do we need to put the invitee in CC for this case?
		SET @add_invitee_in_cc = IF (@invitation_type = 'type_cc'
			, 1
			, 0
			)
			;
	# Change the assignee for the case if needed
	# This procedure needs the following objects:
	#	- variables:
	#		- @bz_user_id
	#		- @creator_bz_id
	#		- @bz_case_id
		CALL `change_case_assignee`;

	# Add the invited user in CC of the case if needed
	# This procedure needs the following objects:
	#	- variables:
		CALL `add_invitee_in_cc`;

# Update the table 'ut_invitation_api_data' so we record what we have done

	# Timestamp	
		SET @timestamp = NOW();
		
	# We do the update to record that we have reached the end of the script...
		UPDATE `ut_invitation_api_data`
			SET `processed_datetime` = @timestamp
				, `script` = @this_script
			WHERE `mefe_invitation_id` = @mefe_invitation_id
			;
