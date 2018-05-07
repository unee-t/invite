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
	SET @environment = 1;
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
# Built for BZFE database v3.0
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
#	- Check if this new user is the first in this role for this unit.
#		- If it IS the first in this role for this unit.
#		 	- Replace the Default 'dummy user' for a specific role with the BZ user in CC for this role for this unit.
#		- If it is NOT the first in this role for this unit.
#			- Do NOT replace the Default assignee for this component/role
#	- Reset the permissions for this unit for this user to the default permissions
#	- WIP Remove this user from the list of user in default CC for a case for this role in this unit.
#
#
#	- Add an existing BZ user as ASSIGNEE to an existing case which has already been created.
#	- Add a comment in the table 'longdesc' to the case to explain that the invitation has been sent to the invited user
#	- Record the change of assignee in the bug activity table so that we have history
#	- Does NOT update the bug_user_last_visit table as the user had no action in there.
#	- Check if the user is a MEFE user only and IF the user is a MEFE user only disable the mail sending functionality from the BZFE.
#
# Limits of this script:
#	- Unit must have all roles created with Dummy user roles.
#
#####################################################
#					
# First we need to define all the variables we need
#					
#####################################################

# Default values:
	
	#User Permissions in the unit:
		SET @can_see_time_tracking = 1;
		SET @can_create_shared_queries = 0;
		SET @can_tag_comment = 1;
		SET @can_create_new_cases = 1;
		SET @can_edit_a_case = 1;
		SET @can_see_all_public_cases = 1;
		SET @can_edit_all_field_in_a_case_regardless_of_role = 1;
		SET @can_see_unit_in_search = 1;
		SET @user_is_publicly_visible = 1;
		SET @user_can_see_publicly_visible = 1;
		SET @can_ask_to_approve_flags = 1;
		SET @can_approve_all_flags = 1;
	
	# Do we need to make the invitee a default CC for all new cases for this role in this unit?
		SET @user_in_default_cc_for_cases = 0;

# Info about this script
	SET @this_script = 'process_one_invitation_all_scenario_v3.0.sql';

# Timestamp	
	SET @timestamp = NOW();
	
# We create a temporary table to record the ids of the dummy users in each environments:
	/*Table structure for table `ut_temp_dummy_users_for_roles` */
		DROP TABLE IF EXISTS `ut_temp_dummy_users_for_roles`;

		CREATE TABLE `ut_temp_dummy_users_for_roles` (
		  `environment_id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Id of the environment',
		  `environment_name` varchar(256) COLLATE utf8_unicode_ci NOT NULL,
		  `tenant_id` int(11) NOT NULL,
		  `landlord_id` int(11) NOT NULL,
		  `contractor_id` int(11) NOT NULL,
		  `mgt_cny_id` int(11) NOT NULL,
		  `agent_id` int(11) DEFAULT NULL,
		  PRIMARY KEY (`environment_id`)
		) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

	/*Data for the table `ut_temp_dummy_users_for_roles` */
		INSERT INTO `ut_temp_dummy_users_for_roles`(`environment_id`,`environment_name`,`tenant_id`,`landlord_id`,`contractor_id`,`mgt_cny_id`,`agent_id`) values 
			(1,'DEV/Staging',96,94,93,95,92),
			(2,'Prod',93,91,90,92,89),
			(3,'demo/dev',4,3,5,6,2);
	
# The reference of the record we want to update in the table ''
	SET @reference_for_update = (SELECT `id` FROM `ut_invitation_api_data` WHERE `mefe_invitation_id` = @mefe_invitation_id);	

# The MEFE information:
	SET @mefe_invitor_user_id = (SELECT `mefe_invitor_user_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

# The unit name and description
	SET @product_id = (SELECT `bz_unit_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

# The Invitor - BZ user id of the user that has genereated the invitation.
	SET @creator_bz_id = (SELECT `bzfe_invitor_user_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

	# We populate the additional variables that we will need for this script to work:
		SET @creator_pub_name = (SELECT `realname` FROM `profiles` WHERE `userid` = @creator_bz_id);
	
# The user who you want to associate to this unit - BZ user id of the user that you want to associate/invite to the unit.
	SET @bz_user_id = (SELECT `bz_user_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

	# We populate the additional variables that we will need for this script to work:
		SET @role_user_g_description = (SELECT `role_type` FROM `ut_role_types` WHERE `id_role_type`=@id_role_type);
		SET @user_pub_name = (SELECT (LEFT(`login_name`,INSTR(`login_name`,"@")-1)) FROM `profiles` WHERE `userid` = @bz_user_id);
		SET @role_user_pub_info = CONCAT(@user_pub_name
								, IF (@role_user_more = '', '', ' - ')
								, IF (@role_user_more = '', '', @role_user_more)
								)
								;
		SET @user_role_desc = (CONCAT(@role_user_g_description, ' - ',@role_user_pub_info));
	
# Role in this unit for the invited user:
	#	- Tenant 1
	# 	- Landlord 2
	#	- Agent 5
	#	- Contractor 3
	#	- Management company 4
	SET @id_role_type = (SELECT `user_role_type_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	SET @role_user_more = (SELECT `user_more` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	SET @user_role_type_description = (SELECT `bz_description` FROM `ut_role_types` WHERE `id_role_type` = @id_role_type);
	SET @user_role_type_name = (SELECT `role_type` FROM `ut_role_types` WHERE `id_role_type` = @id_role_type);
	
	# We need to get the component_id for this role for this product/unit
	# We get that from the ut_product_group table.
		SET @component_id_this_role = (SELECT `component_id` 
									FROM `ut_product_group` 
									WHERE `product_id` = @product_id 
										AND `role_type_id` = @id_role_type
										AND `group_type_id` = 2)
										;
				
	# Is the current assignee for this role for this unit one of the dummy user in this environment?

		# What is the CURRENT default assignee for the role this user has been invited to?
			SET @current_default_assignee_this_role = (SELECT `initialowner` FROM `components` WHERE `id` = @component_id_this_role);

		# What is the default dummy user id for this environment?
		
			# Get the BZ profile id of the dummy users based on the environment variable
				# Tenant 1
					SET @bz_user_id_dummy_tenant = (SELECT `tenant_id` 
												FROM `ut_temp_dummy_users_for_roles` 
												WHERE `environment_id` = @environment)
												;

				# Landlord 2
					SET @bz_user_id_dummy_landlord = (SELECT `landlord_id` 
												FROM `ut_temp_dummy_users_for_roles` 
												WHERE `environment_id` = @environment)
												;
					
				# Contractor 3
					SET @bz_user_id_dummy_contractor = (SELECT `contractor_id` 
												FROM `ut_temp_dummy_users_for_roles` 
												WHERE `environment_id` = @environment)
												;
					
				# Management company 4
					SET @bz_user_id_dummy_mgt_cny = (SELECT `mgt_cny_id` 
												FROM `ut_temp_dummy_users_for_roles` 
												WHERE `environment_id` = @environment)
												;
					
				# Agent 5
					SET @bz_user_id_dummy_agent = (SELECT `agent_id` 
												FROM `ut_temp_dummy_users_for_roles` 
												WHERE `environment_id` = @environment)
												;

		# What is the BZ dummy user id for this role in this script?
			SET @bz_user_id_dummy_user_this_role = IF( @id_role_type = 1
											, @bz_user_id_dummy_tenant
											, IF (@id_role_type = 2
												, @bz_user_id_dummy_landlord
												, IF (@id_role_type = 3
													, @bz_user_id_dummy_contractor
													, IF (@id_role_type = 4
														, @bz_user_id_dummy_mgt_cny
														, IF (@id_role_type = 5
															, @bz_user_id_dummy_agent
															, 'Something is very wrong!!'
															)
														)
													)
												)
											)
											;

# Answer to the question "Is the current default assignee for this role one of the dummy users?"
	SET @is_current_assignee_this_role_a_dummy_user = IF (@current_default_assignee_this_role = @bz_user_id_dummy_user_this_role
		, 1
		, 0
		)
		;	
	
# Is the invited user an occupant of the unit?
	SET @is_occupant = (SELECT `is_occupant` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);

# The case id
	SET @bz_case_id = (SELECT `bz_case_id` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	
# What type of invitation is this?
	SET @invitation_type = (SELECT `invitation_type` FROM `ut_invitation_api_data` WHERE `id` = @reference_for_update);
	
# Do we need to disable the BZ email notification for this user?
	SET @is_mefe_only_user = (SELECT `is_mefe_only_user` 
							FROM `ut_invitation_api_data` 
							WHERE `id` = @reference_for_update)
							;

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
							
#################################################################
#
# All the variables have been set - we can call the procedures
#
#################################################################

# This is legacy and we keep this alive for now:
# The user

	# We record the information about the users that we have just created
	# If this is the first time we record something for this user for this unit, we create a new record.
	# If there is already a record for THAT USER for THIS, then we are updating the information
		
		INSERT INTO `ut_map_user_unit_details`
			(`created`
			, `record_created_by`
			, `user_id`
			, `bz_profile_id`
			, `bz_unit_id`
			, `role_type_id`
			, `can_see_time_tracking`
			, `can_create_shared_queries`
			, `can_tag_comment`
			, `is_occupant`
			, `is_public_assignee`
			, `is_see_visible_assignee`
			, `is_in_cc_for_role`
			, `can_create_case`
			, `can_edit_case`
			, `can_see_case`
			, `can_edit_all_field_regardless_of_role`
			, `is_flag_requestee`
			, `is_flag_approver`
			, `can_create_any_sh`
			, `can_create_same_sh`
			, `can_approve_user_for_flags`
			, `can_decide_if_user_visible`
			, `can_decide_if_user_can_see_visible`
			, `public_name`
			, `more_info`
			, `comment`
			)
			VALUES
			(@timestamp
			, @creator_bz_id
			, @bz_user_id
			, @bz_user_id
			, @product_id
			, @id_role_type
			# Global permission for the whole installation
			, @can_see_time_tracking
			, @can_create_shared_queries
			, @can_tag_comment
			# Attributes of the user
			, @is_occupant
			# User visibility
			, @user_is_publicly_visible
			, @user_can_see_publicly_visible
			# Permissions for cases for this unit.
			, @add_invitee_in_cc
			, @can_create_new_cases
			, @can_edit_a_case
			, @can_see_all_public_cases
			, @can_edit_all_field_in_a_case_regardless_of_role
			# For the flags
			, @can_ask_to_approve_flags
			, @can_approve_all_flags
			# Permissions to create or modify other users
			, 0
			, 0
			, 0
			, 0
			, 0
			, @user_pub_name
			, @role_user_more
			, CONCAT('On '
					, @timestamp
					, ': Created with the script - '
					, @script
					, '.\r\ '
					, `comment`)
			)
			ON DUPLICATE KEY UPDATE
			`created` = @timestamp
			, `record_created_by` = @creator_bz_id
			, `role_type_id` = @id_role_type
			, `can_see_time_tracking` = @can_see_time_tracking
			, `can_create_shared_queries` = @can_create_shared_queries
			, `can_tag_comment` = @can_tag_comment
			, `is_occupant` = @is_occupant
			, `is_public_assignee` = @user_is_publicly_visible
			, `is_see_visible_assignee` = @user_can_see_publicly_visible
			, `is_in_cc_for_role` = @add_invitee_in_cc
			, `can_create_case` = @can_create_new_cases
			, `can_edit_case` = @can_edit_a_case
			, `can_see_case` = @can_see_all_public_cases
			, `can_edit_all_field_regardless_of_role` = @can_edit_all_field_in_a_case_regardless_of_role
			, `is_flag_requestee` = @can_ask_to_approve_flags
			, `is_flag_approver` = @can_approve_all_flags
			, `can_create_any_sh` = 0
			, `can_create_same_sh` = 0
			, `can_approve_user_for_flags` = 0
			, `can_decide_if_user_visible` = 0
			, `can_decide_if_user_can_see_visible` = 0
			, `public_name` = @user_pub_name
			, `more_info` = CONCAT('On: ', @timestamp, '.\r\Updated to ', @role_user_more, '. \r\ ', `more_info`)
			, `comment` = CONCAT('On ', @timestamp, '.\r\Updated with the script - ', @script, '.\r\ ', `comment`)
		;

# Create the table to prepare the permissions
	CALL `create_temp_table_to_update_permissions`;
	
# Revoke all permissions for this user in this unit
	CALL `revoke_all_permission_for_this_user_in_this_unit`;
	
# Prepare the permissions:
	# Generic Permissions
		# These need the following objects:
		#	- table 'ut_user_group_map_temp'
		#	- Variables:
		#		- @bz_user_id
			CALL `can_see_time_tracking`;
			CALL `can_create_shared_queries`;
			CALL `can_tag_comment`;
	# Product/Unit specific permissions
		# These need the following objects:
		#	- table 'ut_user_group_map_temp'
		#	- Variables:
		#		- @bz_user_id
		#		- @product_id
			CALL `can_create_new_cases`;
			CALL `can_edit_a_case`;
			CALL `can_see_all_public_cases`;
			CALL `can_edit_all_field_in_a_case_regardless_of_role`;
			CALL `can_see_unit_in_search`;
			
			CALL `user_is_publicly_visible`;
			CALL `user_can_see_publicly_visible`;
			
			CALL `can_ask_to_approve_flags`;
			CALL `can_approve_all_flags`;
	# Role/Component specific permissions
		# These need the following objects:
		#	- table 'ut_user_group_map_temp'
		#	- Variables:
		#		- @id_role_type
		#		- @bz_user_id
		#		- @product_id
			CALL `show_to_tenant`;
			CALL `is_tenant`;
			CALL `default_tenant_can_see_tenant`;
			
			CALL `show_to_landlord`;
			CALL `are_users_landlord`;
			CALL `default_landlord_see_users_landlord`;
			
			CALL `show_to_contractor`;
			CALL `are_users_contractor`;
			CALL `default_contractor_see_users_contractor`;
			
			CALL `show_to_mgt_cny`;
			CALL `are_users_mgt_cny`;
			CALL `default_mgt_cny_see_users_mgt_cny`;
			
			CALL `show_to_agent`;
			CALL `are_users_agent`;
			CALL `default_agent_see_users_agent`;
			
			CALL `show_to_occupant`;
			CALL `is_occupant`;
			CALL `default_occupant_can_see_occupant`;
	
# All the permission have been prepared, we can now update the permissions table
#		- This NEEDS the table 'ut_user_group_map_temp'
	CALL `update_permissions_invited_user`;
	
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
	
# Replace the default user for this role if needed
# This procedure needs the following objects:
#	- variables:
#		- @is_current_assignee_this_role_a_dummy_user
#		- @component_id_this_role
#		- @bz_user_id
#		- @user_role_desc
#		- @id_role_type
#		- @user_pub_name
#		- @product_id
#		- @creator_bz_id
#		- @mefe_invitation_id
#		- @mefe_invitor_user_id
#		- @is_occupant
#		- @is_mefe_only_user
#		- @role_user_more
	CALL `update_assignee_if_dummy_user`;

# Disable the BZ email notification engine
# This procedure needs the following objects:
#	- variables:
#		- @creator_bz_id
#		- @bz_user_id
	CALL `disable_bugmail`;

# Make the invited user default CC for all cases in this unit if needed
# This procedure needs the following objects:
#	- variables:
#		- @bz_user_id
#		- @product_id
#		- @component_id
#		- @role_user_g_description
	# Make sure the variable we need is correctly defined
		SET @component_id = @component_id_this_role;
	
	# Run the procedure
		CALL `user_in_default_cc_for_cases`;

# Update the table 'ut_data_to_add_user_to_a_case' so that we record what we have done

	# Timestamp	
		SET @timestamp = NOW();

	# Now we can do the update
		INSERT INTO `ut_data_to_add_user_to_a_case`
			( `mefe_invitation_id`
			, `mefe_invitor_user_id`
			, `bzfe_invitor_user_id`
			, `bz_user_id`
			, `bz_case_id`
			, `bz_created_date`
			, `comment`
			)
		VALUES
			(@mefe_invitation_id
			, @mefe_invitor_user_id
			, @creator_bz_id
			, @bz_user_id
			, @bz_case_id
			, @timestamp
			, CONCAT ('inserted in BZ with the script \''
					, @this_script
					, '\'\r\ '
					, IFNULL(`comment`, '')
					)
			)
			;

# Update the table 'ut_invitation_api_data' so we record what we have done

	# Timestamp	
		SET @timestamp = NOW();
		
	# We do the update to record that we have reached the end of the script...
		UPDATE `ut_invitation_api_data`
			SET `processed_datetime` = @timestamp
				, `script` = @this_script
			WHERE `mefe_invitation_id` = @mefe_invitation_id
			;
			
#Clean up
		
	# We Delete the temp table as we do not need it anymore
		DROP TABLE IF EXISTS `ut_temp_dummy_users_for_roles`;
