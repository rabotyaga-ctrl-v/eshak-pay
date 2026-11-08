-- DROP SCHEMA public;

CREATE SCHEMA public AUTHORIZATION pg_database_owner;
-- public.users определение

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE public.users (
	id uuid NOT NULL,
	email text NOT NULL,
	password_hash text NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	CONSTRAINT users_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id)
);


-- public.accounts определение

-- Drop table

-- DROP TABLE public.accounts;

CREATE TABLE public.accounts (
	id uuid NOT NULL,
	user_id uuid NULL,
	account_type text NOT NULL,
	is_primary bool DEFAULT false NOT NULL,
	currency bpchar(3) DEFAULT 'RUB'::bpchar NOT NULL,
	status text DEFAULT 'ACTIVE'::text NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	closed_at timestamptz NULL,
	CONSTRAINT accounts_account_type_check CHECK ((account_type = ANY (ARRAY['USER'::text, 'SYSTEM'::text]))),
	CONSTRAINT accounts_owner_type_check CHECK ((((account_type = 'USER'::text) AND (user_id IS NOT NULL)) OR ((account_type = 'SYSTEM'::text) AND (user_id IS NULL)))),
	CONSTRAINT accounts_pkey PRIMARY KEY (id),
	CONSTRAINT accounts_status_check CHECK ((status = ANY (ARRAY['ACTIVE'::text, 'CLOSED'::text]))),
	CONSTRAINT chk_account_closed_at CHECK ((((status = 'ACTIVE'::text) AND (closed_at IS NULL)) OR ((status = 'CLOSED'::text) AND (closed_at IS NOT NULL)))),
	CONSTRAINT chk_account_owner CHECK ((((account_type = 'USER'::text) AND (user_id IS NOT NULL)) OR ((account_type = 'SYSTEM'::text) AND (user_id IS NULL)))),
	CONSTRAINT chk_system_not_primary CHECK (((account_type <> 'SYSTEM'::text) OR (is_primary = false))),
	CONSTRAINT accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
);
CREATE INDEX idx_accounts_user_id ON public.accounts USING btree (user_id);
CREATE UNIQUE INDEX uq_accounts_primary_per_user ON public.accounts USING btree (user_id) WHERE ((is_primary = true) AND (account_type = 'USER'::text) AND (status = 'ACTIVE'::text));


-- public.cards определение

-- Drop table

-- DROP TABLE public.cards;

CREATE TABLE public.cards (
	id uuid NOT NULL,
	account_id uuid NOT NULL,
	"token" text NOT NULL,
	encrypted_number bytea NOT NULL,
	pan_fingerprint bytea NOT NULL,
	last_four bpchar(4) NOT NULL,
	status text DEFAULT 'ACTIVE'::text NOT NULL,
	reissued_from_card_id uuid NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	blocked_at timestamptz NULL,
	closed_at timestamptz NULL,
	CONSTRAINT cards_pan_fingerprint_key UNIQUE (pan_fingerprint),
	CONSTRAINT cards_pkey PRIMARY KEY (id),
	CONSTRAINT cards_status_check CHECK ((status = ANY (ARRAY['ACTIVE'::text, 'BLOCKED'::text, 'CLOSED'::text]))),
	CONSTRAINT cards_token_key UNIQUE (token),
	CONSTRAINT chk_card_status_dates CHECK ((((status = 'ACTIVE'::text) AND (blocked_at IS NULL) AND (closed_at IS NULL)) OR ((status = 'BLOCKED'::text) AND (blocked_at IS NOT NULL) AND (closed_at IS NULL)) OR ((status = 'CLOSED'::text) AND (closed_at IS NOT NULL)))),
	CONSTRAINT cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id),
	CONSTRAINT cards_reissued_from_card_id_fkey FOREIGN KEY (reissued_from_card_id) REFERENCES public.cards(id)
);
CREATE INDEX idx_cards_account_id ON public.cards USING btree (account_id);


-- public.sessions определение

-- Drop table

-- DROP TABLE public.sessions;

CREATE TABLE public.sessions (
	id uuid NOT NULL,
	user_id uuid NOT NULL,
	token_hash text NOT NULL,
	expires_at timestamptz NOT NULL,
	revoked_at timestamptz NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	CONSTRAINT sessions_pkey PRIMARY KEY (id),
	CONSTRAINT sessions_token_hash_key UNIQUE (token_hash),
	CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
);
CREATE INDEX idx_sessions_expires_at ON public.sessions USING btree (expires_at);
CREATE INDEX idx_sessions_user_id ON public.sessions USING btree (user_id);


-- public.transactions определение

-- Drop table

-- DROP TABLE public.transactions;

CREATE TABLE public.transactions (
	id uuid NOT NULL,
	initiated_by_user_id uuid NULL,
	source_account_id uuid NOT NULL,
	destination_account_id uuid NOT NULL,
	"type" text NOT NULL,
	amount_kopecks int8 NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	CONSTRAINT chk_transaction_different_accounts CHECK ((source_account_id <> destination_account_id)),
	CONSTRAINT transactions_amount_kopecks_check CHECK ((amount_kopecks > 0)),
	CONSTRAINT transactions_pkey PRIMARY KEY (id),
	CONSTRAINT transactions_type_check CHECK ((type = ANY (ARRAY['INITIAL_CREDIT'::text, 'TRANSFER'::text]))),
	CONSTRAINT transactions_destination_account_id_fkey FOREIGN KEY (destination_account_id) REFERENCES public.accounts(id),
	CONSTRAINT transactions_initiated_by_user_id_fkey FOREIGN KEY (initiated_by_user_id) REFERENCES public.users(id),
	CONSTRAINT transactions_source_account_id_fkey FOREIGN KEY (source_account_id) REFERENCES public.accounts(id)
);
CREATE INDEX idx_transactions_destination_account ON public.transactions USING btree (destination_account_id, created_at DESC);
CREATE INDEX idx_transactions_source_account ON public.transactions USING btree (source_account_id, created_at DESC);
CREATE UNIQUE INDEX uq_initial_credit_per_account ON public.transactions USING btree (destination_account_id) WHERE (type = 'INITIAL_CREDIT'::text);


-- public.transfer_requests определение

-- Drop table

-- DROP TABLE public.transfer_requests;

CREATE TABLE public.transfer_requests (
	id uuid NOT NULL,
	user_id uuid NOT NULL,
	idempotency_key text NOT NULL,
	request_hash text NOT NULL,
	status text NOT NULL,
	rejection_reason text NULL,
	transaction_id uuid NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	completed_at timestamptz NULL,
	CONSTRAINT chk_transfer_request_result CHECK ((((status = 'PROCESSING'::text) AND (transaction_id IS NULL) AND (rejection_reason IS NULL) AND (completed_at IS NULL)) OR ((status = 'EXECUTED'::text) AND (transaction_id IS NOT NULL) AND (rejection_reason IS NULL) AND (completed_at IS NOT NULL)) OR ((status = 'REJECTED'::text) AND (transaction_id IS NULL) AND (rejection_reason IS NOT NULL) AND (completed_at IS NOT NULL)))),
	CONSTRAINT transfer_requests_pkey PRIMARY KEY (id),
	CONSTRAINT transfer_requests_status_check CHECK ((status = ANY (ARRAY['PROCESSING'::text, 'EXECUTED'::text, 'REJECTED'::text]))),
	CONSTRAINT transfer_requests_transaction_id_key UNIQUE (transaction_id),
	CONSTRAINT uq_transfer_request_key UNIQUE (user_id, idempotency_key),
	CONSTRAINT transfer_requests_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id),
	CONSTRAINT transfer_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
);
CREATE INDEX idx_transfer_requests_user_created ON public.transfer_requests USING btree (user_id, created_at DESC);


-- public.postings определение

-- Drop table

-- DROP TABLE public.postings;

CREATE TABLE public.postings (
	id uuid NOT NULL,
	transaction_id uuid NOT NULL,
	account_id uuid NOT NULL,
	amount_kopecks int8 NOT NULL,
	created_at timestamptz DEFAULT now() NOT NULL,
	CONSTRAINT postings_amount_kopecks_check CHECK ((amount_kopecks <> 0)),
	CONSTRAINT postings_pkey PRIMARY KEY (id),
	CONSTRAINT postings_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id),
	CONSTRAINT postings_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transactions(id)
);
CREATE INDEX idx_postings_account_id ON public.postings USING btree (account_id, created_at DESC);
CREATE INDEX idx_postings_transaction_id ON public.postings USING btree (transaction_id);



-- DROP FUNCTION public.get_transaction_history(uuid, int4);

CREATE OR REPLACE FUNCTION public.get_transaction_history(p_user_id uuid, p_limit integer DEFAULT 20)
 RETURNS TABLE(transaction_id uuid, transaction_type text, amount_kopecks bigint, created_at timestamp with time zone)
 LANGUAGE sql
AS $function$
    SELECT
        t.id,
        t.type::TEXT,
        p.amount_kopecks,
        p.created_at
    FROM postings p
    JOIN transactions t
        ON p.transaction_id = t.id
    JOIN accounts a
        ON a.id = p.account_id
    WHERE a.user_id = p_user_id
    ORDER BY p.created_at DESC
    FETCH FIRST p_limit ROWS ONLY;
$function$
;