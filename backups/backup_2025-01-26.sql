--
-- PostgreSQL database dump
--

-- Dumped from database version 16.6 (Ubuntu 16.6-1.pgdg22.04+1)
-- Dumped by pg_dump version 16.6 (Ubuntu 16.6-1.pgdg22.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admins; Type: TABLE; Schema: public; Owner: godb
--

CREATE TABLE public.admins (
    id bigint NOT NULL
);


ALTER TABLE public.admins OWNER TO godb;

--
-- Name: channels; Type: TABLE; Schema: public; Owner: godb
--

CREATE TABLE public.channels (
    name character varying(50)
);


ALTER TABLE public.channels OWNER TO godb;

--
-- Name: configs; Type: TABLE; Schema: public; Owner: godb
--

CREATE TABLE public.configs (
    bot_token character varying(100)
);


ALTER TABLE public.configs OWNER TO godb;

--
-- Name: users; Type: TABLE; Schema: public; Owner: godb
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    lang text,
    status integer DEFAULT 1,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.users OWNER TO godb;

--
-- Data for Name: admins; Type: TABLE DATA; Schema: public; Owner: godb
--

COPY public.admins (id) FROM stdin;
6358749851
\.


--
-- Data for Name: channels; Type: TABLE DATA; Schema: public; Owner: godb
--

COPY public.channels (name) FROM stdin;
GrScan
\.


--
-- Data for Name: configs; Type: TABLE DATA; Schema: public; Owner: godb
--

COPY public.configs (bot_token) FROM stdin;
8049808973:AAFcOHe-g2CmGUq65wGhvKlsmvbOE1v5DFs
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: godb
--

COPY public.users (id, lang, status, created_at) FROM stdin;
6358749851	\N	1	2025-01-26 19:58:25.332421
6059047907	\N	1	2025-01-26 20:04:34.073864
458786736	\N	1	2025-01-26 20:15:54.568046
\.


--
-- Name: admins admins_id_key; Type: CONSTRAINT; Schema: public; Owner: godb
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_id_key UNIQUE (id);


--
-- Name: users users_id_key; Type: CONSTRAINT; Schema: public; Owner: godb
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_id_key UNIQUE (id);


--
-- PostgreSQL database dump complete
--

