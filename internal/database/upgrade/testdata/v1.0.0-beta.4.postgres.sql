--
-- PostgreSQL database dump
--

\restrict tekZw2TQqGUBNw2gPJC7ZqGH8ghaRWhK6BuBqfgibgng3tMrp2xdahMQHyAdsk2

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_table_access_method = heap;

--
-- Name: announcements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.announcements (
    id text NOT NULL,
    severity text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    dismissible boolean NOT NULL,
    active boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_keys (
    id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id text NOT NULL,
    name text NOT NULL,
    key_hash text NOT NULL,
    key_prefix text NOT NULL,
    scope text DEFAULT 'read'::text NOT NULL,
    expires_at timestamp with time zone,
    last_used_at timestamp with time zone,
    last_used_ip text DEFAULT ''::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    CONSTRAINT api_keys_scope_check CHECK ((scope = ANY (ARRAY['read'::text, 'read_write'::text])))
);


--
-- Name: component_notification_channels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_notification_channels (
    component_id text NOT NULL,
    notification_channel_id text CONSTRAINT component_notification_channel_notification_channel_id_not_null NOT NULL
);


--
-- Name: components; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.components (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    description text,
    last_notification_status text DEFAULT 'up'::text NOT NULL,
    grouping_window_seconds integer DEFAULT 30 NOT NULL
);


--
-- Name: dashboards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dashboards (
    id text NOT NULL,
    owner_id text NOT NULL,
    name text NOT NULL,
    scope jsonb NOT NULL,
    widgets jsonb NOT NULL,
    default_time_range text NOT NULL,
    refresh_interval text NOT NULL,
    visibility text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: escalation_policies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.escalation_policies (
    id text NOT NULL,
    name text NOT NULL,
    scope_kind text NOT NULL,
    scope_value text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    priority integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: escalation_steps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.escalation_steps (
    id text NOT NULL,
    policy_id text NOT NULL,
    step_order integer NOT NULL,
    delay_minutes integer NOT NULL,
    channel_ids jsonb DEFAULT '[]'::jsonb NOT NULL
);


--
-- Name: expiry_notification_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.expiry_notification_logs (
    id text NOT NULL,
    resource_id text NOT NULL,
    expiry_type text NOT NULL,
    threshold integer NOT NULL,
    sent_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT expiry_notification_logs_expiry_type_check CHECK ((expiry_type = ANY (ARRAY['ssl'::text, 'domain'::text])))
);


--
-- Name: host_alert_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.host_alert_state (
    host_id text NOT NULL,
    state text NOT NULL,
    offline_since timestamp with time zone,
    alerted boolean NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: host_credentials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.host_credentials (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    host_id text NOT NULL,
    hash text NOT NULL,
    prefix text NOT NULL,
    is_active boolean NOT NULL,
    last_used_at timestamp with time zone
);


--
-- Name: host_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.host_metrics (
    id text NOT NULL,
    host_id text NOT NULL,
    sampled_at timestamp with time zone NOT NULL,
    cpu_pct double precision NOT NULL,
    mem_pct double precision NOT NULL,
    net_in bigint NOT NULL,
    net_out bigint NOT NULL,
    disks jsonb
);


--
-- Name: hosts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.hosts (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    os text,
    agent_version text,
    last_seen_at timestamp with time zone,
    last_cpu_pct double precision,
    last_mem_pct double precision,
    last_disk_pct double precision,
    last_net_in bigint,
    last_net_out bigint,
    last_disks jsonb
);


--
-- Name: incident_diagnostics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_diagnostics (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    incident_id text NOT NULL,
    request_method text DEFAULT ''::text NOT NULL,
    request_url text DEFAULT ''::text NOT NULL,
    request_headers text DEFAULT '{}'::text NOT NULL,
    request_timeout integer DEFAULT 0 NOT NULL,
    http_status_code integer DEFAULT '-1'::integer NOT NULL,
    response_headers text DEFAULT '{}'::text NOT NULL,
    response_body text DEFAULT ''::text NOT NULL,
    response_size integer DEFAULT 0 NOT NULL,
    failure_type text DEFAULT ''::text NOT NULL,
    error_message text DEFAULT ''::text NOT NULL,
    error_summary text DEFAULT ''::text NOT NULL,
    total_duration integer DEFAULT 0 NOT NULL,
    dns_duration integer DEFAULT 0 NOT NULL,
    tls_duration integer DEFAULT 0 NOT NULL,
    first_byte_duration integer DEFAULT 0 NOT NULL,
    body_truncated boolean DEFAULT false NOT NULL,
    body_encoded boolean DEFAULT false NOT NULL,
    icmp_available boolean,
    icmp_reachable boolean,
    icmp_rtt_ms integer,
    root_cause_hint text DEFAULT ''::text NOT NULL,
    keyword text,
    keyword_mode text,
    keyword_found boolean
);


--
-- Name: incident_event_steps; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_event_steps (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    incident_id text NOT NULL,
    step text NOT NULL,
    message text
);


--
-- Name: incident_updates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_updates (
    id text NOT NULL,
    incident_id text NOT NULL,
    status text NOT NULL,
    message text DEFAULT ''::text NOT NULL,
    posted_by text DEFAULT ''::text NOT NULL,
    posted_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: incidents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incidents (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_id text NOT NULL,
    cause text DEFAULT 'unknown_failure'::text NOT NULL,
    resolved_at timestamp with time zone,
    started_at timestamp with time zone NOT NULL,
    details bytea
);


--
-- Name: maintenance_resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.maintenance_resources (
    maintenance_id text NOT NULL,
    resource_id text NOT NULL
);


--
-- Name: maintenances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.maintenances (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    title text NOT NULL,
    description text,
    strategy text NOT NULL,
    status text DEFAULT ''::text NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    cron_expr text,
    window_minutes integer,
    timezone text,
    effective_from timestamp with time zone,
    effective_until timestamp with time zone,
    started_at timestamp with time zone,
    ended_at timestamp with time zone
);


--
-- Name: monitoring_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.monitoring_activities (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_id text NOT NULL,
    message text NOT NULL,
    success boolean DEFAULT false NOT NULL,
    response_time integer DEFAULT 0 NOT NULL,
    response_data bytea,
    is_maintenance boolean DEFAULT false NOT NULL
);


--
-- Name: notification_channels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_channels (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    config bytea NOT NULL,
    enabled_by_default boolean DEFAULT false NOT NULL,
    last_sent_at timestamp with time zone,
    last_failure_at timestamp with time zone,
    failures_24h integer DEFAULT 0 NOT NULL
);


--
-- Name: notification_escalation_state; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_escalation_state (
    id text NOT NULL,
    last_digest_at timestamp with time zone,
    watermark_occurred_at timestamp with time zone,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: notification_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notification_events (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    incident_id text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    claim_owner text,
    claimed_at timestamp with time zone,
    processed_at timestamp with time zone,
    last_error text DEFAULT ''::text NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id text NOT NULL,
    user_id text,
    category text NOT NULL,
    severity text NOT NULL,
    title text NOT NULL,
    description text,
    deep_link text,
    payload jsonb,
    occurred_at timestamp with time zone NOT NULL,
    read_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: report_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report_history (
    id text NOT NULL,
    period text NOT NULL,
    sent_at timestamp with time zone NOT NULL,
    status text NOT NULL,
    uptime_pct double precision NOT NULL,
    incident_count integer NOT NULL,
    downtime_seconds bigint NOT NULL,
    recipient_email text NOT NULL,
    resource_breakdown jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: report_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report_settings (
    id text NOT NULL,
    enabled boolean NOT NULL,
    recipient_email text NOT NULL,
    schedule text NOT NULL,
    scope text NOT NULL,
    last_sent_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: resource_credentials; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.resource_credentials (
    id character varying(26) NOT NULL,
    resource_id character varying(26) NOT NULL,
    username character varying(128),
    password bytea NOT NULL,
    options bytea,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: resource_notification_channels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.resource_notification_channels (
    resource_id text NOT NULL,
    notification_channel_id text NOT NULL
);


--
-- Name: resource_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.resource_tags (
    resource_id text NOT NULL,
    tag_id text NOT NULL
);


--
-- Name: resources; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.resources (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    "interval" integer DEFAULT 300 NOT NULL,
    timeout integer DEFAULT 10 NOT NULL,
    target text NOT NULL,
    last_checked timestamp with time zone,
    status text DEFAULT 'pending'::text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    failure_count integer DEFAULT 0 NOT NULL,
    ssl_expiration_date timestamp with time zone,
    ssl_issuer text,
    domain_expiration_date timestamp with time zone,
    domain_registrar text,
    component_id text,
    confirmation_checks integer DEFAULT 2 NOT NULL,
    confirmation_interval integer DEFAULT 30 NOT NULL,
    expiry_alert_thresholds text,
    flap_detection_enabled boolean DEFAULT true NOT NULL,
    flap_threshold integer DEFAULT 4 NOT NULL,
    flap_window_seconds integer DEFAULT 600 NOT NULL,
    flap_max_duration_minutes integer DEFAULT 30 NOT NULL,
    last_status_transition timestamp with time zone,
    flap_started_at timestamp with time zone,
    reminder_interval_minutes integer DEFAULT 0 NOT NULL,
    heartbeat_slug text,
    heartbeat_interval integer,
    heartbeat_grace integer,
    last_ping_at timestamp without time zone,
    keyword text,
    keyword_mode text,
    protocol_type text,
    protocol_port integer,
    host_id text
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version text NOT NULL,
    name text NOT NULL,
    applied_at timestamp with time zone NOT NULL
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id text NOT NULL,
    user_id text NOT NULL,
    browser text DEFAULT ''::text NOT NULL,
    os text DEFAULT ''::text NOT NULL,
    ip text DEFAULT ''::text NOT NULL,
    location text,
    last_active_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone
);


--
-- Name: status_page_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.status_page_settings (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text DEFAULT 'Status Page'::text NOT NULL,
    homepage_url text DEFAULT ''::text NOT NULL,
    custom_domain text DEFAULT ''::text NOT NULL,
    google_analytics_id text DEFAULT ''::text NOT NULL,
    enable_details_page boolean DEFAULT true NOT NULL,
    show_uptime_percentage boolean DEFAULT true NOT NULL,
    hide_paused_monitors boolean DEFAULT true NOT NULL,
    show_incident_history boolean DEFAULT true NOT NULL,
    custom_domain_status text DEFAULT 'pending'::text NOT NULL,
    custom_domain_ssl_status text DEFAULT 'none'::text NOT NULL,
    custom_domain_dns_records jsonb DEFAULT '[]'::jsonb NOT NULL,
    logo_url_light text DEFAULT ''::text NOT NULL,
    logo_url_dark text DEFAULT ''::text NOT NULL,
    favicon_url text DEFAULT ''::text NOT NULL,
    primary_color text DEFAULT '#4f46e5'::text NOT NULL,
    theme_overrides jsonb DEFAULT '{}'::jsonb NOT NULL,
    umami_website_id text DEFAULT ''::text NOT NULL,
    umami_script_url text DEFAULT ''::text NOT NULL
);


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tags (
    id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name text NOT NULL,
    color text,
    description text
);


--
-- Name: two_factor_reset_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.two_factor_reset_tokens (
    token_hash text NOT NULL,
    user_id text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: uptime_daily_agg; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.uptime_daily_agg (
    resource_id text NOT NULL,
    day date NOT NULL,
    samples integer DEFAULT 0 NOT NULL,
    up integer DEFAULT 0 NOT NULL,
    degraded integer DEFAULT 0 NOT NULL,
    down integer DEFAULT 0 NOT NULL,
    uptime_ratio numeric(5,4) DEFAULT 1.0000 NOT NULL,
    computed_at timestamp with time zone NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id text NOT NULL,
    email text NOT NULL,
    name text DEFAULT ''::text NOT NULL,
    hashed_password text NOT NULL,
    password_initialized boolean DEFAULT false NOT NULL,
    force_password_change boolean DEFAULT false NOT NULL,
    two_factor_enabled boolean DEFAULT false NOT NULL,
    two_factor_secret text DEFAULT ''::text NOT NULL,
    two_factor_backup_codes bytea,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Data for Name: announcements; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: api_keys; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: component_notification_channels; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: components; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: dashboards; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: escalation_policies; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: escalation_steps; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: expiry_notification_logs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: host_alert_state; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: host_credentials; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.host_credentials VALUES ('01M2FQTVXFB065TXN8P21BFBMM', '2026-09-14 10:36:33.071414+00', '2026-09-14 10:36:33.071414+00', '01M2FQTVXCR61K1SHHXCMT82RX', '334bf68481a1f1ead22e4fc9a34de0b69da565de076ffca7b4be2e2b5d5d11a0', 'ag_live_932a', true, NULL);


--
-- Data for Name: host_metrics; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: hosts; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.hosts VALUES ('01M2FQTVXCR61K1SHHXCMT82RX', '2026-09-14 10:36:33.068469+00', '2026-09-14 10:36:33.068469+00', 'b4-host', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);


--
-- Data for Name: incident_diagnostics; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.incident_diagnostics VALUES ('01M2FQVJ3FVNGBFXGAZXHRZAY3', '2026-09-14 10:36:55.791061+00', '2026-09-14 10:36:55.803805+00', '01M2FQVJ3CK39T0NVWH3A8KH2C', 'HEAD', 'http://127.0.0.1:1/', '{}', 2, -1, '{}', '', 0, 'HTTP Request Failed', 'request error: Head "http://127.0.0.1:1/": safenet: failed to connect to any resolved IP for 127.0.0.1:1', '', 0, 0, 0, 0, false, false, true, true, 0, 'service_down', NULL, NULL, NULL);


--
-- Data for Name: incident_event_steps; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.incident_event_steps VALUES ('01M2FQVJ3J8KP1QGBVGXQM015F', '2026-09-14 10:36:55.794663+00', '2026-09-14 10:36:55.794663+00', '01M2FQVJ3CK39T0NVWH3A8KH2C', 'detected', 'Incident detected: HTTP request failed');


--
-- Data for Name: incident_updates; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.incident_updates VALUES ('01M2FQVJ3N8GW9249PPJH8JW6P', '01M2FQVJ3CK39T0NVWH3A8KH2C', 'investigating', 'We are currently investigating this issue.', '', '2026-09-14 10:36:55.797248+00', '2026-09-14 10:36:55.797249+00', '2026-09-14 10:36:55.797249+00');


--
-- Data for Name: incidents; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.incidents VALUES ('01M2FQVJ3CK39T0NVWH3A8KH2C', '2026-09-14 10:36:55.788733+00', '2026-09-14 10:36:55.788733+00', '01M2FQTVY2Z9DKV1CCQ85G22WM', 'HTTP Request Failed', NULL, '2026-09-14 10:36:55.788705+00', '\x72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31');


--
-- Data for Name: maintenance_resources; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: maintenances; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: monitoring_activities; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.monitoring_activities VALUES ('01M2FQV7B8R2QQNQV60GH7JWCC', '2026-09-14 10:36:44.77695+00', '2026-09-14 10:36:44.77695+00', '01M2FQTVY2Z9DKV1CCQ85G22WM', 'Check http - Status: down', false, 1, '\x72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31', false);
INSERT INTO public.monitoring_activities VALUES ('01M2FQVJ34SJ5KC3NZVJ0NFP7A', '2026-09-14 10:36:55.780783+00', '2026-09-14 10:36:55.780783+00', '01M2FQTVY2Z9DKV1CCQ85G22WM', 'Check http - Status: down', false, 1, '\x72657175657374206572726f723a20486561642022687474703a2f2f3132372e302e302e313a312f223a20736166656e65743a206661696c656420746f20636f6e6e65637420746f20616e79207265736f6c76656420495020666f72203132372e302e302e313a31', false);


--
-- Data for Name: notification_channels; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: notification_escalation_state; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: notification_events; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.notifications VALUES ('01M2FQVJ3FZJ2EBV72TJCJ49RT', NULL, 'incident', 'error', 'b4-mon is down', NULL, '/incidents/01M2FQVJ3CK39T0NVWH3A8KH2C', '{"incident_id": "01M2FQVJ3CK39T0NVWH3A8KH2C", "resource_id": "01M2FQTVY2Z9DKV1CCQ85G22WM"}', '2026-09-14 10:36:55.791027+00', NULL, '2026-09-14 10:36:55.791114+00', '2026-09-14 10:36:55.791114+00');


--
-- Data for Name: report_history; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: report_settings; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: resource_credentials; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: resource_notification_channels; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: resource_tags; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.resource_tags VALUES ('01M2FQTVY2Z9DKV1CCQ85G22WM', '01M2FQTVY0XMA6FKZFKVQE5V5Y');


--
-- Data for Name: resources; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.resources VALUES ('01M2FQTVZ09319JN3D5KQW4PXA', '2026-09-14 10:36:33.120603+00', '2026-09-14 10:36:33.120603+00', 'b4-mon-up', 'http', 60, 5, 'http://127.0.0.1:18094/api/v1/openapi.json', NULL, 'pending', true, 0, NULL, '', NULL, '', NULL, 2, 30, NULL, true, 4, 600, 30, NULL, NULL, 0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.resources VALUES ('01M2FQTVY2Z9DKV1CCQ85G22WM', '2026-09-14 10:36:33.090979+00', '2026-09-14 10:36:33.110313+00', 'b4-mon', 'http', 10, 2, 'http://127.0.0.1:1/', '2026-09-14 10:36:55.785091+00', 'down', true, 2, NULL, '', NULL, '', NULL, 2, 9, NULL, true, 4, 600, 30, '2026-09-14 10:36:44.785137+00', NULL, 0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '01M2FQTVXCR61K1SHHXCMT82RX');


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.schema_migrations VALUES ('0000', 'schema_migrations', '2026-09-14 10:36:21.582882+00');
INSERT INTO public.schema_migrations VALUES ('0001', 'initial', '2026-09-14 10:36:21.610779+00');
INSERT INTO public.schema_migrations VALUES ('0002', 'indexes', '2026-09-14 10:36:21.631299+00');
INSERT INTO public.schema_migrations VALUES ('0003', 'confirmation_fields', '2026-09-14 10:36:21.634397+00');
INSERT INTO public.schema_migrations VALUES ('0004', 'expiry_notification_logs', '2026-09-14 10:36:21.63928+00');
INSERT INTO public.schema_migrations VALUES ('0005', 'intelligent_alerting_fields', '2026-09-14 10:36:21.644221+00');
INSERT INTO public.schema_migrations VALUES ('0006', 'live_monitor_indexes', '2026-09-14 10:36:21.646621+00');
INSERT INTO public.schema_migrations VALUES ('0007', 'api_keys', '2026-09-14 10:36:21.651697+00');
INSERT INTO public.schema_migrations VALUES ('0008', 'pending_notification_retry', '2026-09-14 10:36:21.654754+00');
INSERT INTO public.schema_migrations VALUES ('0009', 'icmp_diagnostics_fields', '2026-09-14 10:36:21.656835+00');
INSERT INTO public.schema_migrations VALUES ('0010', 'heartbeat_fields', '2026-09-14 10:36:21.660166+00');
INSERT INTO public.schema_migrations VALUES ('0011', 'keyword_fields', '2026-09-14 10:36:21.662496+00');
INSERT INTO public.schema_migrations VALUES ('0012', 'protocol_fields', '2026-09-14 10:36:21.664509+00');
INSERT INTO public.schema_migrations VALUES ('0013', 'resource_credentials', '2026-09-14 10:36:21.668608+00');
INSERT INTO public.schema_migrations VALUES ('0014', 'sessions.down', '2026-09-14 10:36:21.670229+00');
INSERT INTO public.schema_migrations VALUES ('0015', 'two_factor.down', '2026-09-14 10:36:21.67599+00');
INSERT INTO public.schema_migrations VALUES ('0016', 'escalation_policies.down', '2026-09-14 10:36:21.681715+00');
INSERT INTO public.schema_migrations VALUES ('0017', 'custom_domains.down', '2026-09-14 10:36:21.689165+00');
INSERT INTO public.schema_migrations VALUES ('0018', 'statuspage_domain_state.down', '2026-09-14 10:36:21.694301+00');
INSERT INTO public.schema_migrations VALUES ('0019', 'statuspage_branding.down', '2026-09-14 10:36:21.700455+00');
INSERT INTO public.schema_migrations VALUES ('0020', 'uptime_daily_agg.down', '2026-09-14 10:36:21.704595+00');
INSERT INTO public.schema_migrations VALUES ('0021', 'incident_updates.down', '2026-09-14 10:36:21.709765+00');
INSERT INTO public.schema_migrations VALUES ('0022', 'notification_channel_dispatch_stats.down', '2026-09-14 10:36:21.714807+00');
INSERT INTO public.schema_migrations VALUES ('0023', 'replace_ga_with_umami.down', '2026-09-14 10:36:21.718121+00');
INSERT INTO public.schema_migrations VALUES ('0024', 'notifications.down', '2026-09-14 10:36:21.72138+00');
INSERT INTO public.schema_migrations VALUES ('0025', 'dashboards.down', '2026-09-14 10:36:21.726897+00');
INSERT INTO public.schema_migrations VALUES ('0026', 'report_history.down', '2026-09-14 10:36:21.732136+00');
INSERT INTO public.schema_migrations VALUES ('0027', 'announcements.down', '2026-09-14 10:36:21.739792+00');
INSERT INTO public.schema_migrations VALUES ('0028', 'hosts.down', '2026-09-14 10:36:21.744805+00');
INSERT INTO public.schema_migrations VALUES ('0029', 'host_alert_state.down', '2026-09-14 10:36:21.753975+00');
INSERT INTO public.schema_migrations VALUES ('0030', 'notification_escalation_state.down', '2026-09-14 10:36:21.759148+00');


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.sessions VALUES ('01M2FQTVWSE6W3FVS92A7NPGCC', '01M2FQTH21T09D3DE1HMD318A1', 'Unknown', 'Unknown', '[::1]:51274', NULL, '2026-09-14 10:36:57.837854+00', '2026-09-14 10:36:33.049332+00', NULL);


--
-- Data for Name: status_page_settings; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: tags; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.tags VALUES ('01M2FQTVY0XMA6FKZFKVQE5V5Y', '2026-09-14 10:36:33.088294+00', '2026-09-14 10:36:33.088294+00', 'b4', NULL, NULL);


--
-- Data for Name: two_factor_reset_tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: uptime_daily_agg; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.users VALUES ('01M2FQTH21T09D3DE1HMD318A1', 'admin@ogoune.test', 'Administrator', '$2a$12$J8I.pohkRdV1r1Hu8xbOj.d22unOxgc8SkljzE/tbbIhCEVCb7Cqu', true, false, false, '', NULL, '2026-09-14 10:36:33.06594+00', '2026-09-14 10:36:21.953307+00', '2026-09-14 10:36:21.953307+00');


--
-- Name: announcements announcements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.announcements
    ADD CONSTRAINT announcements_pkey PRIMARY KEY (id);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: component_notification_channels component_notification_channels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_notification_channels
    ADD CONSTRAINT component_notification_channels_pkey PRIMARY KEY (component_id, notification_channel_id);


--
-- Name: components components_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.components
    ADD CONSTRAINT components_pkey PRIMARY KEY (id);


--
-- Name: dashboards dashboards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboards
    ADD CONSTRAINT dashboards_pkey PRIMARY KEY (id);


--
-- Name: escalation_policies escalation_policies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.escalation_policies
    ADD CONSTRAINT escalation_policies_pkey PRIMARY KEY (id);


--
-- Name: escalation_steps escalation_steps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.escalation_steps
    ADD CONSTRAINT escalation_steps_pkey PRIMARY KEY (id);


--
-- Name: expiry_notification_logs expiry_notification_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.expiry_notification_logs
    ADD CONSTRAINT expiry_notification_logs_pkey PRIMARY KEY (id);


--
-- Name: expiry_notification_logs expiry_notification_logs_resource_id_expiry_type_threshold_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.expiry_notification_logs
    ADD CONSTRAINT expiry_notification_logs_resource_id_expiry_type_threshold_key UNIQUE (resource_id, expiry_type, threshold);


--
-- Name: host_alert_state host_alert_state_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.host_alert_state
    ADD CONSTRAINT host_alert_state_pkey PRIMARY KEY (host_id);


--
-- Name: host_credentials host_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.host_credentials
    ADD CONSTRAINT host_credentials_pkey PRIMARY KEY (id);


--
-- Name: host_metrics host_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.host_metrics
    ADD CONSTRAINT host_metrics_pkey PRIMARY KEY (id);


--
-- Name: hosts hosts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hosts
    ADD CONSTRAINT hosts_pkey PRIMARY KEY (id);


--
-- Name: incident_diagnostics incident_diagnostics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_diagnostics
    ADD CONSTRAINT incident_diagnostics_pkey PRIMARY KEY (id);


--
-- Name: incident_event_steps incident_event_steps_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_event_steps
    ADD CONSTRAINT incident_event_steps_pkey PRIMARY KEY (id);


--
-- Name: incident_updates incident_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_updates
    ADD CONSTRAINT incident_updates_pkey PRIMARY KEY (id);


--
-- Name: incidents incidents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT incidents_pkey PRIMARY KEY (id);


--
-- Name: maintenance_resources maintenance_resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.maintenance_resources
    ADD CONSTRAINT maintenance_resources_pkey PRIMARY KEY (maintenance_id, resource_id);


--
-- Name: maintenances maintenances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.maintenances
    ADD CONSTRAINT maintenances_pkey PRIMARY KEY (id);


--
-- Name: monitoring_activities monitoring_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monitoring_activities
    ADD CONSTRAINT monitoring_activities_pkey PRIMARY KEY (id);


--
-- Name: notification_channels notification_channels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_channels
    ADD CONSTRAINT notification_channels_pkey PRIMARY KEY (id);


--
-- Name: notification_escalation_state notification_escalation_state_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_escalation_state
    ADD CONSTRAINT notification_escalation_state_pkey PRIMARY KEY (id);


--
-- Name: notification_events notification_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_events
    ADD CONSTRAINT notification_events_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: report_history report_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_history
    ADD CONSTRAINT report_history_pkey PRIMARY KEY (id);


--
-- Name: report_settings report_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report_settings
    ADD CONSTRAINT report_settings_pkey PRIMARY KEY (id);


--
-- Name: resource_credentials resource_credentials_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_credentials
    ADD CONSTRAINT resource_credentials_pkey PRIMARY KEY (id);


--
-- Name: resource_credentials resource_credentials_resource_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_credentials
    ADD CONSTRAINT resource_credentials_resource_id_key UNIQUE (resource_id);


--
-- Name: resource_notification_channels resource_notification_channels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_notification_channels
    ADD CONSTRAINT resource_notification_channels_pkey PRIMARY KEY (resource_id, notification_channel_id);


--
-- Name: resource_tags resource_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_tags
    ADD CONSTRAINT resource_tags_pkey PRIMARY KEY (resource_id, tag_id);


--
-- Name: resources resources_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT resources_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: status_page_settings status_page_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_settings
    ADD CONSTRAINT status_page_settings_pkey PRIMARY KEY (id);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: two_factor_reset_tokens two_factor_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.two_factor_reset_tokens
    ADD CONSTRAINT two_factor_reset_tokens_pkey PRIMARY KEY (token_hash);


--
-- Name: uptime_daily_agg uptime_daily_agg_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.uptime_daily_agg
    ADD CONSTRAINT uptime_daily_agg_pkey PRIMARY KEY (resource_id, day);


--
-- Name: api_keys uq_api_keys_key_hash; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT uq_api_keys_key_hash UNIQUE (key_hash);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_announcements_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_announcements_active ON public.announcements USING btree (active, created_at DESC);


--
-- Name: idx_api_keys_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_keys_active ON public.api_keys USING btree (user_id, is_active);


--
-- Name: idx_api_keys_key_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_keys_key_hash ON public.api_keys USING btree (key_hash);


--
-- Name: idx_api_keys_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_api_keys_user_id ON public.api_keys USING btree (user_id);


--
-- Name: idx_component_notification_channels_channel_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_notification_channels_channel_id ON public.component_notification_channels USING btree (notification_channel_id);


--
-- Name: idx_components_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_components_created_at ON public.components USING btree (created_at);


--
-- Name: idx_components_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_components_name ON public.components USING btree (name);


--
-- Name: idx_dashboards_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_dashboards_owner ON public.dashboards USING btree (owner_id);


--
-- Name: idx_dashboards_updated; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_dashboards_updated ON public.dashboards USING btree (updated_at DESC);


--
-- Name: idx_escalation_policies_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_escalation_policies_active ON public.escalation_policies USING btree (is_active, priority);


--
-- Name: idx_expiry_notification_logs_expiry_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expiry_notification_logs_expiry_type ON public.expiry_notification_logs USING btree (expiry_type);


--
-- Name: idx_expiry_notification_logs_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expiry_notification_logs_resource_id ON public.expiry_notification_logs USING btree (resource_id);


--
-- Name: idx_expiry_notification_logs_sent_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_expiry_notification_logs_sent_at ON public.expiry_notification_logs USING btree (sent_at);


--
-- Name: idx_host_credentials_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_host_credentials_hash ON public.host_credentials USING btree (hash);


--
-- Name: idx_host_credentials_host_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_host_credentials_host_id ON public.host_credentials USING btree (host_id);


--
-- Name: idx_host_metrics_host_sampled; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_host_metrics_host_sampled ON public.host_metrics USING btree (host_id, sampled_at);


--
-- Name: idx_hosts_last_seen_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_hosts_last_seen_at ON public.hosts USING btree (last_seen_at);


--
-- Name: idx_incident_diagnostics_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_diagnostics_created_at ON public.incident_diagnostics USING btree (created_at);


--
-- Name: idx_incident_diagnostics_failure_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_diagnostics_failure_type ON public.incident_diagnostics USING btree (failure_type);


--
-- Name: idx_incident_diagnostics_http_status_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_diagnostics_http_status_code ON public.incident_diagnostics USING btree (http_status_code);


--
-- Name: idx_incident_diagnostics_incident_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_incident_diagnostics_incident_id ON public.incident_diagnostics USING btree (incident_id);


--
-- Name: idx_incident_event_steps_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_event_steps_created_at ON public.incident_event_steps USING btree (created_at);


--
-- Name: idx_incident_event_steps_incident_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_event_steps_incident_id ON public.incident_event_steps USING btree (incident_id);


--
-- Name: idx_incident_event_steps_step; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_event_steps_step ON public.incident_event_steps USING btree (step);


--
-- Name: idx_incident_updates_incident_posted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_updates_incident_posted ON public.incident_updates USING btree (incident_id, posted_at DESC);


--
-- Name: idx_incidents_cause; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_cause ON public.incidents USING btree (cause);


--
-- Name: idx_incidents_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_created_at ON public.incidents USING btree (created_at);


--
-- Name: idx_incidents_resolved_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_resolved_at ON public.incidents USING btree (resolved_at);


--
-- Name: idx_incidents_resource_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_resource_active ON public.incidents USING btree (resource_id, started_at DESC) WHERE (resolved_at IS NULL);


--
-- Name: idx_incidents_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_resource_id ON public.incidents USING btree (resource_id);


--
-- Name: idx_incidents_started_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_started_at ON public.incidents USING btree (started_at);


--
-- Name: idx_maintenance_resources_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenance_resources_resource_id ON public.maintenance_resources USING btree (resource_id);


--
-- Name: idx_maintenances_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_created_at ON public.maintenances USING btree (created_at);


--
-- Name: idx_maintenances_cron_expr; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_cron_expr ON public.maintenances USING btree (cron_expr);


--
-- Name: idx_maintenances_effective_from; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_effective_from ON public.maintenances USING btree (effective_from);


--
-- Name: idx_maintenances_effective_until; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_effective_until ON public.maintenances USING btree (effective_until);


--
-- Name: idx_maintenances_end_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_end_at ON public.maintenances USING btree (end_at);


--
-- Name: idx_maintenances_ended_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_ended_at ON public.maintenances USING btree (ended_at);


--
-- Name: idx_maintenances_start_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_start_at ON public.maintenances USING btree (start_at);


--
-- Name: idx_maintenances_started_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_started_at ON public.maintenances USING btree (started_at);


--
-- Name: idx_maintenances_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_status ON public.maintenances USING btree (status);


--
-- Name: idx_maintenances_strategy; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_maintenances_strategy ON public.maintenances USING btree (strategy);


--
-- Name: idx_monitoring_activities_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monitoring_activities_created_at ON public.monitoring_activities USING btree (created_at);


--
-- Name: idx_monitoring_activities_resource_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monitoring_activities_resource_created ON public.monitoring_activities USING btree (resource_id, created_at);


--
-- Name: idx_monitoring_activities_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_monitoring_activities_resource_id ON public.monitoring_activities USING btree (resource_id);


--
-- Name: idx_notification_channels_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_channels_created_at ON public.notification_channels USING btree (created_at);


--
-- Name: idx_notification_channels_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_channels_type ON public.notification_channels USING btree (type);


--
-- Name: idx_notification_events_claim_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_events_claim_owner ON public.notification_events USING btree (claim_owner);


--
-- Name: idx_notification_events_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_events_created_at ON public.notification_events USING btree (created_at);


--
-- Name: idx_notification_events_incident_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_events_incident_id ON public.notification_events USING btree (incident_id);


--
-- Name: idx_notification_events_status_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_events_status_created_at ON public.notification_events USING btree (status, created_at);


--
-- Name: idx_notification_events_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notification_events_type ON public.notification_events USING btree (type);


--
-- Name: idx_notifications_occurred; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_occurred ON public.notifications USING btree (occurred_at DESC);


--
-- Name: idx_notifications_user_occurred; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_occurred ON public.notifications USING btree (user_id, occurred_at DESC);


--
-- Name: idx_report_history_period; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_report_history_period ON public.report_history USING btree (period);


--
-- Name: idx_report_history_sent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_report_history_sent ON public.report_history USING btree (sent_at DESC);


--
-- Name: idx_resource_credentials_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resource_credentials_resource_id ON public.resource_credentials USING btree (resource_id);


--
-- Name: idx_resource_notification_channels_channel_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resource_notification_channels_channel_id ON public.resource_notification_channels USING btree (notification_channel_id);


--
-- Name: idx_resource_tags_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resource_tags_tag_id ON public.resource_tags USING btree (tag_id);


--
-- Name: idx_resources_component_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_component_id ON public.resources USING btree (component_id);


--
-- Name: idx_resources_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_created_at ON public.resources USING btree (created_at);


--
-- Name: idx_resources_heartbeat_missed; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_heartbeat_missed ON public.resources USING btree (type, status, last_ping_at) WHERE ((type = 'heartbeat'::text) AND (is_active = true));


--
-- Name: idx_resources_heartbeat_slug_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_resources_heartbeat_slug_unique ON public.resources USING btree (heartbeat_slug) WHERE (heartbeat_slug IS NOT NULL);


--
-- Name: idx_resources_host_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_host_id ON public.resources USING btree (host_id);


--
-- Name: idx_resources_last_transition; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_last_transition ON public.resources USING btree (last_status_transition);


--
-- Name: idx_resources_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_status ON public.resources USING btree (status);


--
-- Name: idx_resources_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_resources_type ON public.resources USING btree (type);


--
-- Name: idx_sessions_user_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user_active ON public.sessions USING btree (user_id, revoked_at);


--
-- Name: idx_status_page_settings_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_page_settings_created_at ON public.status_page_settings USING btree (created_at);


--
-- Name: idx_tags_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tags_created_at ON public.tags USING btree (created_at);


--
-- Name: idx_tags_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tags_name ON public.tags USING btree (name);


--
-- Name: idx_two_factor_reset_tokens_expires; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_two_factor_reset_tokens_expires ON public.two_factor_reset_tokens USING btree (expires_at);


--
-- Name: idx_two_factor_reset_tokens_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_two_factor_reset_tokens_user ON public.two_factor_reset_tokens USING btree (user_id);


--
-- Name: idx_uptime_daily_agg_day; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_uptime_daily_agg_day ON public.uptime_daily_agg USING btree (day);


--
-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_created_at ON public.users USING btree (created_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: uniq_escalation_priority_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_escalation_priority_active ON public.escalation_policies USING btree (priority) WHERE (is_active = true);


--
-- Name: uniq_escalation_steps_order; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uniq_escalation_steps_order ON public.escalation_steps USING btree (policy_id, step_order);


--
-- Name: api_keys api_keys_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: dashboards dashboards_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboards
    ADD CONSTRAINT dashboards_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: escalation_steps escalation_steps_policy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.escalation_steps
    ADD CONSTRAINT escalation_steps_policy_id_fkey FOREIGN KEY (policy_id) REFERENCES public.escalation_policies(id) ON DELETE CASCADE;


--
-- Name: expiry_notification_logs expiry_notification_logs_resource_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.expiry_notification_logs
    ADD CONSTRAINT expiry_notification_logs_resource_id_fkey FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: component_notification_channels fk_component_notification_channels_channel; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_notification_channels
    ADD CONSTRAINT fk_component_notification_channels_channel FOREIGN KEY (notification_channel_id) REFERENCES public.notification_channels(id) ON DELETE CASCADE;


--
-- Name: component_notification_channels fk_component_notification_channels_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_notification_channels
    ADD CONSTRAINT fk_component_notification_channels_component FOREIGN KEY (component_id) REFERENCES public.components(id) ON DELETE CASCADE;


--
-- Name: incident_diagnostics fk_incident_diagnostics_incident; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_diagnostics
    ADD CONSTRAINT fk_incident_diagnostics_incident FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;


--
-- Name: incident_event_steps fk_incident_event_steps_incident; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_event_steps
    ADD CONSTRAINT fk_incident_event_steps_incident FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;


--
-- Name: incidents fk_incidents_resource; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT fk_incidents_resource FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: maintenance_resources fk_maintenance_resources_maintenance; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.maintenance_resources
    ADD CONSTRAINT fk_maintenance_resources_maintenance FOREIGN KEY (maintenance_id) REFERENCES public.maintenances(id) ON DELETE CASCADE;


--
-- Name: maintenance_resources fk_maintenance_resources_resource; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.maintenance_resources
    ADD CONSTRAINT fk_maintenance_resources_resource FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: monitoring_activities fk_monitoring_activities_resource; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monitoring_activities
    ADD CONSTRAINT fk_monitoring_activities_resource FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: notification_events fk_notification_events_incident; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notification_events
    ADD CONSTRAINT fk_notification_events_incident FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE;


--
-- Name: resource_notification_channels fk_resource_notification_channels_channel; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_notification_channels
    ADD CONSTRAINT fk_resource_notification_channels_channel FOREIGN KEY (notification_channel_id) REFERENCES public.notification_channels(id) ON DELETE CASCADE;


--
-- Name: resource_notification_channels fk_resource_notification_channels_resource; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_notification_channels
    ADD CONSTRAINT fk_resource_notification_channels_resource FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: resource_tags fk_resource_tags_resource; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_tags
    ADD CONSTRAINT fk_resource_tags_resource FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: resource_tags fk_resource_tags_tag; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_tags
    ADD CONSTRAINT fk_resource_tags_tag FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;


--
-- Name: resources fk_resources_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resources
    ADD CONSTRAINT fk_resources_component FOREIGN KEY (component_id) REFERENCES public.components(id) ON DELETE SET NULL;


--
-- Name: host_alert_state host_alert_state_host_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.host_alert_state
    ADD CONSTRAINT host_alert_state_host_id_fkey FOREIGN KEY (host_id) REFERENCES public.hosts(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: resource_credentials resource_credentials_resource_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.resource_credentials
    ADD CONSTRAINT resource_credentials_resource_id_fkey FOREIGN KEY (resource_id) REFERENCES public.resources(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: two_factor_reset_tokens two_factor_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.two_factor_reset_tokens
    ADD CONSTRAINT two_factor_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict tekZw2TQqGUBNw2gPJC7ZqGH8ghaRWhK6BuBqfgibgng3tMrp2xdahMQHyAdsk2

