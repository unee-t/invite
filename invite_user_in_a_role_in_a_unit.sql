# For any question about this script, ask Franck
#
# This script needs the following pre-requisites:
#	- The data for the invitation are created in the table `ut_invitation_api_data`
#	- The schema of the BZ database is v3.10+
#
#################################################################
#
# UPDATE THE BELOW VARIABLES ACCORDING TO YOUR NEEDS
#
#################################################################
#
# The MEFE invitation id that we want to process:
	SET @mefe_invitation_id = '%s';
	SET @mefe_invitation_id_int_value = %d;
#
# Environment: Which environment are you creating the unit in?
#	- 1 is for the DEV/Staging
#	- 2 is for the prod environment
#	- 3 is for the Demo environment
	SET @environment = %d;
#
# Info about this script
	SET @this_script = 'invite_user_in_a_role_in_a_unit.sql';
#
########################################################################
#
#	ALL THE VARIABLES WE NEED HAVE BEEN DEFINED, WE CAN RUN THE SCRIPT
#
########################################################################
#
#############################################
#
# IMPORTANT INFORMATION ABOUT THIS SCRIPT
#
#############################################
#
# Use this script only if the Unit EXIST in the BZFE 
# It assumes that the unit has been created with all the necessary BZ objects and all the roles assigned to dummy users.
#
# Pre-requisite:
#	- The table 'ut_invitation_api_data' has been updated 
# 	- We know the MEFE Invitation id that we need to process.
#	- We know the environment where this script is run
# 
	CALL `add_user_to_role_in_unit`;
