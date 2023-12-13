-- Table: public.persons

DROP TABLE IF EXISTS public.persons;

CREATE TABLE public.persons
(
    id integer NOT NULL,
    parent_id integer,
    email character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    dob date,
    dateofdeath date,
    salutation character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    firstname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    middlenames character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    surname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    othernames character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    -- correspondencebypost boolean NOT NULL,
    -- correspondencebyphone boolean NOT NULL,
    -- correspondencebyemail boolean NOT NULL,
    occupation character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    createddate timestamp(0) without time zone DEFAULT NULL::timestamp without time zone,
    uid character varying(36) COLLATE pg_catalog."default" NOT NULL,
    type character varying(255) COLLATE pg_catalog."default" NOT NULL,
    lpa002signaturedate date,
    dxnumber character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    dxexchange character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    lpapartcsignaturedate date,
    companyname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    companynumber character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    systemstatus boolean DEFAULT true,
    relationshiptodonor character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    isreplacementattorney boolean DEFAULT false,
    istrustcorporation boolean DEFAULT false,
    trustcorporationappointedas character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    signatoryone character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    signatorytwo character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    companyseal character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    previousnames character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    cannotsignform boolean,
    applyingforfeeremission boolean,
    receivingbenefits boolean,
    receiveddamageaward boolean,
    haslowincome boolean,
    signaturedate date,
    haspreviouslpa boolean,
    notesforpreviouslpa character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    sageid character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    noticegivendate date,
    certificateproviderstatementtype character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    statement character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    certificateproviderskills character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    fullname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    caserecnumber character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    clientaccommodation character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    maritalstatus character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    clientstatus character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    statusdate date,
    deputyreferencenumber character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    deputycompliance character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    relationshiptoclient character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    correspondencebywelsh boolean NOT NULL DEFAULT false,
    medicalcondition text COLLATE pg_catalog."default",
    interpreterrequired character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    mobilenumber character varying(15) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    countryofresidence character varying(40) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    newsletter boolean DEFAULT false,
    specialcorrespondencerequirements_audiotape boolean NOT NULL DEFAULT false,
    specialcorrespondencerequirements_largeprint boolean NOT NULL DEFAULT false,
    specialcorrespondencerequirements_hearingimpaired boolean NOT NULL DEFAULT false,
    specialcorrespondencerequirements_spellingofnamerequirescare boolean NOT NULL DEFAULT false,
    client_id integer,
    companyreference character varying(255) COLLATE pg_catalog."default",
    feepayer_id integer,
    isorganisation boolean DEFAULT false,
    correspondencename character varying(200) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    deputystatus character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    casesmanagedashybrid boolean DEFAULT false,
    deputytype character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    supervisioncaseowner_id integer,
    clientsource character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    memorablephrase character varying(30) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    risk_score integer,
    caseactorgroup character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    updateddate timestamp(0) without time zone DEFAULT NULL::timestamp without time zone,
    executor_id integer,
    deputycasrecid integer,
    organisationname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    organisationteamordepartmentname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    deputynumber integer,
    executivecasemanager_id integer,
    deputysubtype character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    firm_id integer,
    CONSTRAINT persons_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

-- Table: public.addresses

DROP TABLE IF EXISTS public.addresses;

CREATE TABLE public.addresses
(
    id integer NOT NULL,
    person_id integer,
    address_lines json,
    town character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    county character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    postcode character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    country character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    type character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    isairmailrequired boolean,
    CONSTRAINT addresses_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

-- Table: public.cases

DROP TABLE IF EXISTS public.cases;

CREATE TABLE public.cases
(
    id integer NOT NULL,
    assignee_id integer,
    donor_id integer,
    correspondent_id integer,
    client_id integer,
    oldcaseid integer,
    applicationtype integer,
    title character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    casetype character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    casesubtype character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    duedate date,
    registrationdate date,
    closeddate date,
    status character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    rejecteddate timestamp(0) without time zone DEFAULT NULL::timestamp without time zone,
    caseattorneysingular boolean NOT NULL DEFAULT false,
    caseattorneyjointlyandseverally boolean NOT NULL DEFAULT false,
    caseattorneyjointly boolean NOT NULL DEFAULT false,
    caseattorneyjointlyandjointlyandseverally boolean NOT NULL DEFAULT false,
    caseattorneyactionadditionalinfo boolean NOT NULL DEFAULT false,
    repeatapplication boolean NOT NULL DEFAULT false,
    repeatapplicationreference character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    uid character varying(36) COLLATE pg_catalog."default" NOT NULL,
    -- type character varying(255) COLLATE pg_catalog."default" NOT NULL,
    ascertained_by integer DEFAULT 1,
    lpadonorsignaturedate date,
    lpadonorsignature boolean,
    donorsignaturewitnessed boolean DEFAULT false,
    lpadonorsignatoryfullname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    donorhaspreviouslpas boolean DEFAULT false,
    previouslpainfo text COLLATE pg_catalog."default",
    applicantsignaturedate date,
    applicantsignatoryfullname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    createddate timestamp(0) without time zone DEFAULT NULL::timestamp without time zone,
    lifesustainingtreatment character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    lifesustainingtreatmentsignaturedatea date,
    lifesustainingtreatmentsignaturedateb date,
    trustcorporationsignedas integer DEFAULT 1,
    hasrelativetonotice boolean DEFAULT false,
    areallattorneysapplyingtoregister boolean DEFAULT false,
    epadonorsignaturedate date,
    epadonornoticegivendate date,
    donorhasotherepas boolean DEFAULT false,
    otherepainfo text COLLATE pg_catalog."default",
    usesnotifiedpersons boolean DEFAULT false,
    nonoticegiven boolean DEFAULT false,
    notifiedpersonpermissionby integer DEFAULT 1,
    paymentbydebitcreditcard integer DEFAULT 0,
    paymentbycheque integer DEFAULT 0,
    wouldliketoapplyforfeeremission integer DEFAULT 0,
    haveappliedforfeeremission integer DEFAULT 0,
    cardpaymentcontact text COLLATE pg_catalog."default",
    howattorneysact character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    howreplacementattorneysact character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    attorneyactdecisions character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    replacementattorneyactdecisions character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    replacementorder text COLLATE pg_catalog."default",
    additionalinfo text COLLATE pg_catalog."default",
    anyotherinfo boolean DEFAULT false,
    additionalinfodonorsignature boolean DEFAULT false,
    additionalinfodonorsignaturedate date,
    paymentid character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    paymentamount character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    paymentdate date,
    paymentremission integer DEFAULT 0,
    paymentexemption integer DEFAULT 0,
    attorneypartydeclaration integer DEFAULT 1,
    attorneyapplicationassertion integer DEFAULT 1,
    attorneymentalactpermission integer DEFAULT 1,
    attorneydeclarationsignaturedate date,
    attorneydeclarationsignature boolean,
    attorneydeclarationsignaturewitness boolean DEFAULT false,
    attorneydeclarationsignatoryfullname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    correspondentcomplianceassertion integer DEFAULT 1,
    notificationdate date,
    dispatchdate date,
    cancellationdate date,
    applicantsdeclaration integer DEFAULT 1,
    applicantsdeclarationsignaturedate date,
    applicationhasrestrictions boolean,
    applicationhasguidance boolean,
    applicationhascharges boolean DEFAULT false,
    certificateprovidersignaturedate date,
    certificateprovidersignature boolean,
    attorneystatementdate date,
    signingonbehalfdate date,
    signingonbehalffullname character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    noticegivendate date,
    orderdate date,
    orderissuedate date,
    securitybond boolean DEFAULT false,
    bondreferencenumber character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    bondvalue double precision,
    orderstatus character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    statusdate date,
    caserecnumber character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    filelocationdescription character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    feepayer_id integer,
    paymentpostponement integer DEFAULT 0,
    invaliddate date,
    withdrawndate date,
    revokeddate date,
    receiptdate timestamp(0) without time zone DEFAULT NULL::timestamp without time zone,
    batchid character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    howdeputyappointed character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    onlinelpaid character varying(12) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    orderstatusnotes text COLLATE pg_catalog."default",
    orderexpirydate date,
    orderclosurereason character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    ordersubtype character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    ordernotes text COLLATE pg_catalog."default",
    updateddate timestamp(0) without time zone DEFAULT NULL::timestamp without time zone,
    clauseexpirydate timestamp(0) without time zone,
    madeactivedate date,
    workeddate date,
    CONSTRAINT cases_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

-- Table: public.person_caseitem

DROP TABLE IF EXISTS public.person_caseitem;

CREATE TABLE public.person_caseitem
(
    person_id integer NOT NULL,
    caseitem_id integer NOT NULL,
    CONSTRAINT person_caseitem_pkey PRIMARY KEY (person_id, caseitem_id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

-- Table: public.phonenumbers

DROP TABLE IF EXISTS public.phonenumbers;

CREATE TABLE public.phonenumbers
(
    id integer NOT NULL,
    person_id integer,
    phone_number character varying(255) COLLATE pg_catalog."default" DEFAULT NULL::character varying,
    -- type character varying(255) COLLATE pg_catalog."default" NOT NULL,
    -- is_default boolean NOT NULL,
    updateddate timestamp(0) without time zone,
    CONSTRAINT phonenumbers_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

-- Table: supervision.firm

DROP TABLE IF EXISTS supervision.firm;

DROP SCHEMA IF EXISTS supervision;

CREATE SCHEMA supervision;

CREATE TABLE supervision.firm
(
    id integer NOT NULL,
    firmname character varying(255) NOT NULL,
    addressline1 character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    addressline2 character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    addressline3 character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    town character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    county character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    postcode character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    phonenumber character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    email character varying(255) COLLATE pg_catalog."default" default NULL::character varying,
    firmnumber integer NOT NULL,
    piireceived date,
    piiexpiry date,
    piiamount numeric(12, 2) default NULL::numeric,
    piirequested date,
    CONSTRAINT firm_pkey PRIMARY KEY (id)
)
    WITH (
        OIDS = FALSE
    )
    TABLESPACE pg_default;
