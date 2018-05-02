SET @foo_bar_invitation_id = 'from a script 2';

INSERT INTO `ut_test_foo_bar`(`foobar_invitation_id`) 
    VALUES 
    (@foo_bar_invitation_id);
