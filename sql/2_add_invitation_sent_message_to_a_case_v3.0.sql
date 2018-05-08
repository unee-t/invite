# For any question about this script, ask Franck
#
#################################################################
#																#
# UPDATE THE BELOW VARIABLES ACCORDING TO YOUR NEEDS			#
#																#
#################################################################
#
# The MEFE invitation id that we want to finalize:
	SET @mefe_invitation_id = '%s';
#
# The Timestamp when the MEFE invitation was sent:
	SET @mefe_invitation_sent = NOW();
#
########################################################################
#
#	ALL THE VARIABLES WE NEED HAVE BEEN DEFINED, WE CAN RUN THE SCRIPT #
#
########################################################################
#
#############################################
#											#
# IMPORTANT INFORMATION ABOUT THIS SCRIPT	#
#											#
#############################################
#
# Built for BZFE database v3.0
#
# This assumes that the invitation has been:
#	- Created in the MEFE
#	- Processed in the BZ DB
#	- Finalized in the MEFE (an email has been sent to the invited user)
#
# Pre-requisite:
#	- The table 'ut_invitation_api_data' has been updated
# 	- We know the MEFE Invitation id that we need to process.
#	- We know the environment where this script is run
# 
# This script depends on one SQL procedure 'finalize_invitation_to_a_case'.
# This procedure needs the following variables:
#	- @mefe_invitation_id
#	- @mefe_invitation_sent
#	- @bz_case_id
#	- @bz_user_id
#	- @creator_bz_id
#	- @user_role_type_name
#	- @mefe_invitor_user_id
#
# This script assumes that:
#	- An invitation has been created in the MEFE
#	- The invitation has been processed in the BZ database
#	- The MEFE has been notified that the invitation has been processed in the BZ database and has sent an email to the invitee.
#
# This script will:
#	- Populate the needed variables
#	- Verify that the invitation has been correctly processed in the BZ database
# 	- Call the procedure to finalize the invitation
#		- Add a comment to the case to indicate that the invitation has been sent
#		- Log its actions
#		- Update the table 'ut_data_to_add_user_to_a_case' for future reference and audit.
#
#
#####################################################
#					
# First we need to define all the variables we need
#					
#####################################################

# Info about this script
	SET @this_script = 'add_invitation_sent_message_to_a_case_v3.0.sql';

# Timestamp	
	SET @timestamp = NOW();
	
# The reference of the record we want to update in the table ''
	SET @reference_for_update = (SELECT `id` FROM `ut_invitation_api_data` WHERE `mefe_invitation_id` = @mefe_invitation_id);	

# The case id
	SET @bz_case_id = (SELECT `bz_case_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	
# The user who you want to associate to this unit - BZ user id of the user that you want to associate/invite to the unit.
	SET @bz_user_id = (SELECT `bz_user_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

# The Invitor - BZ user id of the user that has genereated the invitation.
	SET @creator_bz_id = (SELECT `bzfe_invitor_user_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	
# Role in this unit for the invited user:
	#	- Tenant 1
	# 	- Landlord 2
	#	- Agent 5
	#	- Contractor 3
	#	- Management company 4
	SET @id_role_type = (SELECT `user_role_type_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	SET @user_role_type_name = (SELECT `role_type` FROM `ut_role_types` WHERE `id_role_type` = @id_role_type);
	
# The MEFE information:
	SET @mefe_invitor_user_id = (SELECT `mefe_invitor_user_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	
# What type of invitation is this?
	SET @invitation_type = (SELECT `invitation_type` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
							
#################################################################
#
# All the variables have been set - we can call the procedures
#
#################################################################

# Finalize the invitation
	CALL `finalize_invitation_to_a_case`;

# Update the table 'ut_invitation_api_data' so we record what we have done
		
	# We do the update to record that we have reached the end of the script...
		UPDATE `ut_invitation_api_data`
			SET `api_post_datetime` = @mefe_invitation_sent
				, `script` = @this_script
			WHERE `mefe_invitation_id` = @mefe_invitation_id
			;
